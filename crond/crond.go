package crond

import (
	"cron-s/internal/util"
	"cron-s/tasks"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"os"
	"time"
)

type Crond struct {
	waitGroup util.WaitGroupWrapper

	opts *options
	Log  *log.Logger

	scheduleTaskTick <-chan time.Time
	waitExecTask     chan *tasks.Task
	waitStoreTask    chan *tasks.Store

	raft       *raft.Raft
	httpServer *http.Server
}

func New(opts *options) *Crond {
	return &Crond{
		opts:             opts,
		Log:              log.New(os.Stdout, "", log.LstdFlags),
		scheduleTaskTick: time.Tick(opts.defaultScheduleTaskTick),
		waitExecTask:     make(chan *tasks.Task, 100),
		waitStoreTask:    make(chan *tasks.Store, 100),
	}
}

func (c *Crond) Run() {
	c.Log.Println("[DEBUG] crond: Run")

	c.waitGroup.Wrap(func() {
		for {
			select {
			case <-c.scheduleTaskTick:
				c.scheduleTask()
			case t := <-c.waitExecTask:
				c.Log.Println("[DEBUG] crond: exec Task")

				c.waitGroup.Wrap(func() {
					if c.raft.State() != raft.Leader {
						return
					}

					ss := tasks.NewStore(t)
					ss.NodeId = c.opts.nodeId
					ss.Ip = c.opts.bind
					ss.StartTime = time.Now()
					ss.Result, ss.Err = t.Exec()
					ss.EndTime = time.Now()

					c.waitStoreTask <- ss
				})
			case st := <-c.waitStoreTask:
				c.Log.Printf("[DEBUG] crond: start Store Task, Name %s, Result %s\n", st.Task.Name, st.Result)

				c.waitGroup.Wrap(func() {
					st.Save()
				})
			}
		}
	})

	var err error
	if c.raft, err = c.newRaft(c.opts); err != nil {
		c.Log.Println("[WARN] crond: newRaft err", err)
	}
	if c.opts.join != "" {
		err := c.joinCluster(c.opts)
		if err != nil {
			c.Log.Println("[WARN] crond: joinCluster err", err)
		}
	}

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

func (c *Crond) scheduleTask() {
	if tasks.Len() < 1 {
		return
	}

	now := time.Now()
	top := tasks.Top()
	if top.RunTime.Before(now) || top.RunTime.Equal(now) {
		wet := *top
		c.waitExecTask <- &wet

		top.RunTime = top.CronExpression.Next(now)
		tasks.Fix(0)
	}

	tick := c.opts.defaultScheduleTaskTick
	if tasks.Top().RunTime.Sub(now) < tick {
		tick = tasks.Top().RunTime.Sub(now)
	}
	c.scheduleTaskTick = time.Tick(tick)
}
