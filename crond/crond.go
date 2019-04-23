package crond

import (
	"container/heap"
	"context"
	"cron-s/internal/util"
	"cron-s/store"
	"cron-s/task"
	"encoding/json"
	"fmt"
	"github.com/gorhill/cronexpr"
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
	Log      *log.Logger
	mu       sync.Mutex
	TaskHeap *task.Heap

	scheduleTaskTick <-chan time.Time
	waitExecTask     chan *task.Task
	waitStoreTask    chan *store.Store

	taskEvent chan *task.Event

	httpServer *http.Server

	raft *raft.Raft
}

func New(opts *options) *Crond {
	return &Crond{
		opts:             opts,
		Log:              log.New(os.Stdout, "", log.LstdFlags),
		TaskHeap:         task.NewHeap(),
		scheduleTaskTick: time.Tick(opts.defaultScheduleTaskTick),
		waitExecTask:     make(chan *task.Task, 100),
		waitStoreTask:    make(chan *store.Store, 100),
		taskEvent:        make(chan *task.Event, 100),
	}
}

func (c *Crond) Run() {
	c.Log.Println("[DEBUG] crond: Run")

	c.waitGroup.Wrap(func() {
		for {
			select {
			case <-c.scheduleTaskTick:
				c.scheduleTask()
			case te := <-c.taskEvent:
				c.Log.Println("[DEBUG] crond: task event")

				tes, err := json.Marshal(te)
				if err != nil {
					c.Log.Println("[WARN] crond: task event Marshal err", err)
					return
				}
				af := c.raft.Apply(tes, 3*time.Second)
				if err := af.Error(); err != nil {
					c.Log.Println("[WARN] crond: task event Apply err", err)
					return
				}
			case t := <-c.waitExecTask:
				c.Log.Println("[DEBUG] crond: exec Task")

				c.waitGroup.Wrap(func() {
					c.execTask(t)
				})
			case st := <-c.waitStoreTask:
				c.Log.Printf("[DEBUG] crond: start Store Task, Name %s, Result %s\n", st.Task.Name, st.Result)
			}
		}
	})

	c.waitGroup.Wrap(func() {
		var err error
		c.raft, err = c.newRaft(c.opts)
		if err != nil {
			c.Log.Println("[WARN] crond: newRaft err", err)
		}
	})

	c.waitGroup.Wrap(func() {
		if c.opts.join != "" {
			err := c.joinCluster(c.opts)
			if err != nil {
				c.Log.Println("[WARN] crond: joinCluster err", err)
			}
		}
	})

	c.initHttpServer()
	if err := c.httpServer.ListenAndServe(); err != nil {
		c.Log.Println("[WARN] crond: listen http err", err)
	}
}

func (c *Crond) Exit() {
	c.waitGroup.Wait()
	c.raft.Shutdown()

	c.Log.Println("[DEBUG] crond: exit")
}

func (c *Crond) HandleTaskEvent(e *task.Event) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch e.Cmd {
	case task.ADD:
		e.Task.CronExpression, err = cronexpr.Parse(e.Task.CronLine)
		if err != nil {
			return
		}
		e.Task.RunTime = e.Task.CronExpression.Next(time.Now())

		heap.Push(c.TaskHeap, e.Task)
	case task.DEL:
		for i, t := range *c.TaskHeap {
			if e.Task.Name == t.Name {
				heap.Remove(c.TaskHeap, i)
			}
		}
	}
	return
}

func (c *Crond) scheduleTask() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.TaskHeap.Len() < 1 {
		return
	}

	now := time.Now()
	top := c.TaskHeap.Top().(*task.Task)
	if top.RunTime.Before(now) || top.RunTime.Equal(now) {
		wet := *top
		c.waitExecTask <- &wet

		top.RunTime = top.CronExpression.Next(now)
		heap.Fix(c.TaskHeap, 0)
	}

	tick := c.opts.defaultScheduleTaskTick
	if c.TaskHeap.Top().(*task.Task).RunTime.Sub(now) < tick {
		tick = c.TaskHeap.Top().(*task.Task).RunTime.Sub(now)
	}
	c.scheduleTaskTick = time.Tick(tick)
}

func (c *Crond) execTask(t *task.Task) {
	if c.raft.State() != raft.Leader {
		return
	}

	ss := store.NewStore(t)
	ss.NodeId = c.opts.nodeId
	ss.Ip = c.opts.bind
	ss.StartTime = time.Now()

	cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", ss.Task.Cmd)
	ss.Result, ss.Err = cmd.CombinedOutput()

	ss.EndTime = time.Now()
	c.waitStoreTask <- ss
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

func (c *Crond) getTasks() *task.Heap {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.TaskHeap
}
