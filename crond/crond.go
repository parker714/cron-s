package crond

import (
	"context"
	"cron-s/internal/lg"
	"cron-s/internal/server"
	"cron-s/internal/task"
	"cron-s/internal/util"
	"os/exec"
	"sync"
	"time"
)

type Crond struct {
	Opts *Options

	lg *lg.Lg

	waitGroup util.WaitGroupWrapper
	mu        sync.Mutex

	server          server.Server
	tasks           map[string]*task.Task
	checkTasksAfter <-chan time.Time
	taskModifyEvent chan *task.ModifyEvent
	taskScheduling  chan *task.Schedule
	taskScheduled   chan *task.Schedule
}

func New(opts *Options) *Crond {
	c := &Crond{
		Opts:            opts,
		tasks:           make(map[string]*task.Task),
		taskModifyEvent: make(chan *task.ModifyEvent),
		taskScheduling:  make(chan *task.Schedule),
		taskScheduled:   make(chan *task.Schedule),
		checkTasksAfter: make(<-chan time.Time),
	}

	c.lg = lg.New("[crond]", opts.LogLevel)

	return c
}

func (c *Crond) Main() {
	var err error
	c.lg.Logf(lg.INFO, "Crond Run...")

	c.server, err = server.NewEtcd(c.Opts.EtcdEndpoints)
	if err != nil {
		c.lg.Logf(lg.ERROR, "NewEtcd err:%s, EtcdEndpoints:%s", err, c.Opts.EtcdEndpoints)
		return
	}

	c.tasks, err = c.server.Get(server.JobsKey)
	if err != nil {
		c.lg.Logf(lg.ERROR, "Get %s err: %s", server.JobsKey, err)
		return
	}

	c.waitGroup.Wrap(func() {
		c.server.Watch(c.taskModifyEvent)
	})

	c.checkTasksRunTime()
	c.doTaskChan()
}

func (c *Crond) Exit() {
	if c.server != nil {
		c.server.Close()
	}

	c.waitGroup.Wait()
	c.lg.Logf(lg.INFO, "Exit...")
}

func (c *Crond) doTaskChan() {
	for {
		select {
		case <-c.checkTasksAfter:
			c.lg.Logf(lg.INFO, "checkTasksRunTime")

			c.waitGroup.Wrap(func() {
				c.checkTasksRunTime()
			})
		case t := <-c.taskModifyEvent:
			c.lg.Logf(lg.INFO, "Task modify: %s, type: %d", t.Name, t.Type)

			c.mu.Lock()
			switch t.Type {
			case task.PUT:
				c.tasks[t.Name] = t.Task
			case task.DEL:
				if _, ok := c.tasks[t.Name]; ok {
					delete(c.tasks, t.Name)
				}
			}
			c.mu.Unlock()

			c.checkTasksRunTime()
		case t := <-c.taskScheduling:
			c.lg.Logf(lg.INFO, "Task scheduling:%s", t.Name)

			c.waitGroup.Wrap(func() {
				t.StartTime = time.Now()

				t.Err = c.server.Lock(t.Name, func() {
					cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", t.Task.Cmd)
					t.Result, t.Err = cmd.CombinedOutput()
				})

				t.EndTime = time.Now()
				c.taskScheduled <- t
			})
		case t := <-c.taskScheduled:
			c.lg.Logf(lg.INFO, "Task scheduled:%s, result:%s, err:%v", t.Name, t.Result, t.Err)
		}
	}
}

func (c *Crond) checkTasksRunTime() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	after := 24 * 60 * 60 * time.Second
	for _, t := range c.tasks {
		if t.NextRunTime.Before(now) || t.NextRunTime.Equal(now) {
			ts := &task.Schedule{
				Name: t.Name,
				Task: t,
			}
			c.taskScheduling <- ts
			t.NextRunTime = t.CronExpression.Next(now)
		}

		if t.NextRunTime.Sub(now) < after {
			after = t.NextRunTime.Sub(now)
		}
	}

	c.checkTasksAfter = time.After(after)
}
