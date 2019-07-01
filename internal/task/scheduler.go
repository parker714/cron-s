package task

import (
	"cron-s/pkg/util"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"time"
)

type Scheduler struct {
	WaitGroup util.WaitGroupWrapper

	opt  *Option
	Data *Data
	Raft *raft.Raft

	waitRenewTick <-chan time.Time
	waitExec      chan *Task
	waitStorage   chan *storage
}

func NewScheduler(opt *Option, d *Data, r *raft.Raft) *Scheduler {
	return &Scheduler{
		opt:           opt,
		Data:          d,
		Raft:          r,
		waitRenewTick: time.Tick(opt.TaskRenewTick),
		waitExec:      make(chan *Task, 100),
		waitStorage:   make(chan *storage, 100),
	}
}

func (s *Scheduler) Run() {
	s.WaitGroup.Wrap(func() {
		for {
			select {
			case <-s.waitRenewTick:
				log.Debug("Task scheduler renew")
				s.Renew()
			case t := <-s.waitExec:
				log.Debug("Task scheduler exec", t)

				s.WaitGroup.Wrap(func() {
					if s.Raft.State() != raft.Leader {
						return
					}

					ss := NewStorage(t)
					//ss.NodeId = s.cf.Raft.NodeId
					//ss.Ip = s.cf.Raft.Bind
					ss.StartTime = time.Now()
					ss.Result, ss.Err = t.Exec()
					ss.EndTime = time.Now()

					s.waitStorage <- ss
				})
			case st := <-s.waitStorage:
				log.Debug("Task scheduler store", st)

				s.WaitGroup.Wrap(func() {
					st.Save()
				})
			}
		}
	})
}

func (s *Scheduler) Renew() {
	if s.Data.Len() < 1 {
		return
	}

	now := time.Now()
	top := s.Data.Top()
	if top.RunTime.Before(now) || top.RunTime.Equal(now) {
		wet := *top
		s.waitExec <- &wet

		top.RunTime = top.CronExpression.Next(now)
		s.Data.Fix(0)
	}

	tick := s.opt.TaskRenewTick
	if s.Data.Top().RunTime.Sub(now) < tick {
		tick = s.Data.Top().RunTime.Sub(now)
	}
	s.waitRenewTick = time.Tick(tick)
}
