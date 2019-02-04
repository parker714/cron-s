package crond

import (
	"context"
	"cron-s/internal/job"
	"cron-s/internal/lg"
	"cron-s/internal/server"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Crond struct {
	Opts *Options

	Server server.Server

	jobs             map[string]*job.Job
	jobChangeEvent   chan *job.ChangeEvent
	jobCompleteEvent chan *job.CompleteEvent

	mu sync.Mutex
}

func New(opts *Options) *Crond {
	opts.Logger = log.New(os.Stderr, opts.LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)

	return &Crond{
		Opts:             opts,
		jobs:             make(map[string]*job.Job),
		jobChangeEvent:   make(chan *job.ChangeEvent),
		jobCompleteEvent: make(chan *job.CompleteEvent),
	}
}

func (c *Crond) Main() {
	var err error
	c.logf(lg.INFO, "Cron-s Run...")

	c.Server, err = server.NewEtcd(c.Opts.EtcdEndpoints)
	if err != nil {
		c.logf(lg.ERROR, "NewEtcd %s %s", c.Opts.EtcdEndpoints, err)
		return
	}

	c.jobs, err = c.Server.Get()
	if err != nil {
		c.logf(lg.ERROR, "Get Jobs %s", err)
		return
	}

	go c.Server.Watch(c.jobChangeEvent)
	go c.doJobEvent()

	c.scheduling()
}

func (c *Crond) Exit() {
	c.logf(lg.INFO, "Cron-s Exit...")
}

func (c *Crond) doJobEvent() {
	for {
		select {
		case j := <-c.jobChangeEvent:
			c.mu.Lock()
			switch j.Type {
			case job.PUT:
				c.jobs[j.Key] = j.Job
			case job.DEL:
				if _, ok := c.jobs[j.Key]; ok {
					delete(c.jobs, j.Key)
				}
			}
			c.mu.Unlock()
		case j := <-c.jobCompleteEvent:
			c.logf(lg.INFO, "job complete %v", j)
		}
	}
}

func (c *Crond) scheduling() {
	for {
		now := time.Now()
		c.mu.Lock()
		for key, j := range c.jobs {
			if j.NextRunTime.Before(now) || j.NextRunTime.Equal(now) {
				jce := &job.CompleteEvent{
					Key:       key,
					Job:       j,
					StartTime: now,
				}

				go func() {
					err := c.Server.Lock(key, func() {
						cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", j.Cmd)
						jce.Result, jce.Err = cmd.CombinedOutput()
						jce.EndTime = time.Now()

						c.jobCompleteEvent <- jce
					})

					if err != nil {
						c.logf(lg.ERROR, "scheduling %s %s", key, err)
					}
				}()

				j.NextRunTime = j.CronExpression.Next(now)
			}
		}
		c.mu.Unlock()
	}
}
