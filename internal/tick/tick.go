package tick

import (
	"cron-s/internal/conf"
	"cron-s/internal/data"
	"cron-s/internal/tasks"
	"cron-s/pkg/util"
	"github.com/hashicorp/raft"
	"github.com/prometheus/common/log"
	"time"
)

type Tick struct {
	raft      *raft.Raft
	waitGroup util.WaitGroupWrapper

	scheduleTaskTick <-chan time.Time
	waitExecTask     chan *tasks.Task
	waitStoreTask    chan *tasks.Storage
}

func New(raft *raft.Raft) *Tick {
	return &Tick{
		raft:             raft,
		scheduleTaskTick: time.Tick(conf.DefaultScheduleTaskTick),
		waitExecTask:     make(chan *tasks.Task, 100),
		waitStoreTask:    make(chan *tasks.Storage, 100),
	}
}

func (s *Tick) Run() {
	for {
		select {
		case <-s.scheduleTaskTick:
			s.Renew()
		case t := <-s.waitExecTask:
			log.Debug("schedule: exec Task")

			s.waitGroup.Wrap(func() {
				if s.raft.State() != raft.Leader {
					return
				}

				ss := tasks.NewStorage(t)
				ss.NodeId = conf.NodeId
				ss.Ip = conf.Bind
				ss.StartTime = time.Now()
				ss.Result, ss.Err = t.Exec()
				ss.EndTime = time.Now()

				s.waitStoreTask <- ss
			})
		case st := <-s.waitStoreTask:
			log.Debug("schedule: start Store Task, Name %s, Result %s\n", st.Task.Name, st.Result)

			s.waitGroup.Wrap(func() {
				st.Save()
			})
		}
	}
}

func (s *Tick) Renew() {
	if data.Len() < 1 {
		return
	}

	now := time.Now()
	top := data.Top()
	if top.RunTime.Before(now) || top.RunTime.Equal(now) {
		wet := *top
		s.waitExecTask <- &wet

		top.RunTime = top.CronExpression.Next(now)
		data.Fix(0)
	}

	tick := conf.DefaultScheduleTaskTick
	if data.Top().RunTime.Sub(now) < tick {
		tick = data.Top().RunTime.Sub(now)
	}
	s.scheduleTaskTick = time.Tick(tick)
}
