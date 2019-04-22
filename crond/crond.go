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
		scheduleTaskTick: time.Tick(opts.defaultScheduleTime),
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
				//c.scheduleTask()
			case task := <-c.waitExecTask:
				c.log.Println("[INFO] start Exec Task", task.Name)

				c.waitGroup.Wrap(func() {
					store := NewStore()
					store.task = task
					store.startTime = time.Now()

					cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", store.task.Cmd)
					store.result, store.err = cmd.CombinedOutput()

					store.endTime = time.Now()
					c.waitStoreTask <- store
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
	for {
		if c.taskHeap.Len() < 1 {
			return
		}

		now := time.Now()
		if (*c.taskHeap)[0].runTime.Before(now) || (*c.taskHeap)[0].runTime.Equal(now) {
			c.waitExecTask <- (*c.taskHeap)[0]
			(*c.taskHeap)[0].runTime = (*c.taskHeap)[0].cronExpression.Next(now)
			heap.Fix(c.taskHeap, 0)
		}
	}

	//tick := (*c.taskHeap)[0].runTime.Sub(now)
	//if tick < 0 {
	//	tick = c.opts.defaultScheduleTime
	//}
	//c.scheduleTaskTick = time.Tick(tick)
}
