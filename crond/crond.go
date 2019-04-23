package crond

import (
	"container/heap"
	"context"
	"cron-s/internal/util"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Crond struct {
	waitGroup util.WaitGroupWrapper

	opts     *options
	log      *log.Logger
	mu       sync.Mutex
	taskHeap *taskHeap

	scheduleTaskTick <-chan time.Time
	waitExecTask     chan *task
	waitStoreTask    chan *store

	httpServer *http.Server
}

func New(opts *options) *Crond {
	return &Crond{
		opts:             opts,
		log:              log.New(os.Stdout, "", log.LstdFlags),
		taskHeap:         &taskHeap{},
		scheduleTaskTick: time.Tick(opts.defaultScheduleTaskTick),
		waitExecTask:     make(chan *task, 100),
		waitStoreTask:    make(chan *store, 100),
	}
}

func (c *Crond) Run() {
	c.log.Println("Crond Run")

	c.waitGroup.Wrap(func() {
		for {
			select {
			case <-c.scheduleTaskTick:
				c.log.Println("[INFO] start Schedule Task")

				c.scheduleTask()
			case task := <-c.waitExecTask:
				c.log.Println("[INFO] start Exec Task", task.Name)

				c.waitGroup.Wrap(func() {
					c.execTask(task)
				})
			case store := <-c.waitStoreTask:
				c.log.Printf("[INFO] start Store Task, Name %s, Result %s", store.task.Name, store.result)
			}
		}
	})

	c.initHttpServer()
	if err := c.httpServer.ListenAndServe(); err != nil {
		c.log.Println("listen http err", err)
	}
}

func (c *Crond) Exit() {
	c.waitGroup.Wait()
	c.log.Println("Crond Exit")
}

func (c *Crond) scheduleTask() {
	if c.taskHeap.Len() < 1 {
		return
	}

	now := time.Now()
	top := c.taskHeap.Top().(*task)

	if top.runTime.Before(now) || top.runTime.Equal(now) {
		wet := *top
		c.waitExecTask <- &wet

		top.runTime = top.cronExpression.Next(now)
		heap.Fix(c.taskHeap, 0)
	}

	tick := c.opts.defaultScheduleTaskTick
	if c.taskHeap.Top().(*task).runTime.Sub(now) < tick {
		tick = c.taskHeap.Top().(*task).runTime.Sub(now)
	}
	c.scheduleTaskTick = time.Tick(tick)
}

func (c *Crond) execTask(task *task) {
	wst := NewStore(task)
	wst.startTime = time.Now()

	cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", wst.task.Cmd)
	wst.result, wst.err = cmd.CombinedOutput()

	wst.endTime = time.Now()
	c.waitStoreTask <- wst
}
