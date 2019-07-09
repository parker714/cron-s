package task

import (
	cheap "container/heap"
	"cron-s/pkg/util"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"time"
)

// Scheduler is task scheduler struct
type Scheduler struct {
	WaitGroup util.WaitGroupWrapper

	opt  *Option
	Heap Heap
	Raft *raft.Raft

	renewTick   *time.Ticker
	waitExec    chan *Task
	waitStorage chan *Task
}

// NewScheduler returns scheduler instance
func NewScheduler(opt *Option, h Heap, r *raft.Raft) *Scheduler {
	return &Scheduler{
		opt:         opt,
		Heap:        h,
		Raft:        r,
		renewTick:   time.NewTicker(opt.TaskRenewTick),
		waitExec:    make(chan *Task, 100),
		waitStorage: make(chan *Task, 100),
	}
}

// Run start task schedule
func (s *Scheduler) Run() {
	s.WaitGroup.Wrap(func() {
		for {
			select {
			case <-s.renewTick.C:
				log.Debug("Task scheduler renew")
				s.Renew()
			case t := <-s.waitExec:
				log.Debug("Task scheduler exec", t)

				s.WaitGroup.Wrap(func() {
					if s.Raft.State() != raft.Leader {
						return
					}

					t.ActualStartTime = time.Now()
					if err := t.Exec(); err != nil {
						log.Errorf("exec task %+v ,err %s", t, err)
						return
					}
					t.ActualEndTime = time.Now()

					s.waitStorage <- t
				})
			case t := <-s.waitStorage:
				log.Debug("Task scheduler storage", t)

				s.WaitGroup.Wrap(func() {
					t.Save()
				})
			}
		}
	})
}

// Renew is used to renew schedule task
func (s *Scheduler) Renew() {
	if s.Heap.Len() < 1 {
		return
	}

	now := time.Now()
	top, err := s.Heap.Top()
	if err != nil {
		return
	}
	if top.PlanExecTime.Before(now) || top.PlanExecTime.Equal(now) {
		wet := *top
		s.waitExec <- &wet

		top.PlanExecTime = top.CronExpression.Next(now)
		cheap.Fix(s.Heap, 0)
	}

	tick := s.opt.TaskRenewTick
	top, err = s.Heap.Top()
	if err != nil {
		return
	}
	if top.PlanExecTime.Sub(now) < tick {
		tick = top.PlanExecTime.Sub(now)
	}
	s.renewTick = time.NewTicker(tick)
}
