package crond

import (
	"container/heap"
	"context"
	"cron-s/internal/util"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
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

	taskEvent chan *taskEvent

	httpServer *http.Server

	raft *raft.Raft
}

func New(opts *options) *Crond {
	return &Crond{
		opts:             opts,
		log:              log.New(os.Stdout, "", log.LstdFlags),
		taskHeap:         &taskHeap{},
		scheduleTaskTick: time.Tick(opts.defaultScheduleTaskTick),
		waitExecTask:     make(chan *task, 100),
		waitStoreTask:    make(chan *store, 100),
		taskEvent:        make(chan *taskEvent, 100),
	}
}

func (c *Crond) Run() {
	var err error
	c.log.Println("[INFO] crond Run")

	c.waitGroup.Wrap(func() {
		for {
			select {
			case te := <-c.taskEvent:
				c.log.Println("[INFO] task event")

				tes, err := json.Marshal(te)
				if err != nil {
					fmt.Println("task event failed", err)
					return
				}

				af := c.raft.Apply(tes, 0)
				if err := af.Error(); err != nil {
					fmt.Println("Apply failed", err)
					return
				}

				//c.handleTask(te)
			case <-c.scheduleTaskTick:

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

	c.waitGroup.Wrap(func() {
		c.raft, err = c.newRaft(c.opts)
		if err != nil {
			c.log.Println("[WARN] newRaft", err)
		}
	})

	c.waitGroup.Wrap(func() {
		if c.opts.join != "" {
			err := c.joinCluster(c.opts)
			if err != nil {
				c.log.Println("[WARN] join err", err)
			}
		}
	})

	c.initHttpServer()
	if err := c.httpServer.ListenAndServe(); err != nil {
		c.log.Println("[INFO] listen http err", err)
	}
}

func (c *Crond) Exit() {
	c.waitGroup.Wait()
	c.log.Println("[INFO] crond Exit")
}

func (c *Crond) handleTask(te *taskEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch te.Cmd {
	case TASK_ADD:
		heap.Push(c.taskHeap, te.Task)
	case TASK_DEL:
		for i, task := range *c.taskHeap {
			if te.Task.Name == task.Name {
				heap.Remove(c.taskHeap, i)
			}
		}
	}

	// todo
}

func (c *Crond) scheduleTask() {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	if (c.raft.State() != raft.Leader) {
		return
	}

	wst := NewStore(task)
	wst.startTime = time.Now()

	cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", wst.task.Cmd)
	wst.result, wst.err = cmd.CombinedOutput()

	wst.endTime = time.Now()
	c.waitStoreTask <- wst
}

func (c *Crond) joinCluster(opts *options) error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", opts.join, opts.nodeId, opts.bind)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("addCluster url %s err %s", url, err)
	}

	return nil
}
