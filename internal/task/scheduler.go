package task

import (
	"container/heap"
	"context"
	"cron-s/pkg/util"
	"time"

	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

// Scheduler is task scheduler struct
type Scheduler struct {
	WaitGroup util.WaitGroupWrapper

	opt   *Option
	Tasks Tasks
	Raft  *raft.Raft

	renewTick   *time.Ticker
	waitExec    chan *Task
	waitStorage chan *Task
}

// NewScheduler returns scheduler instance
func NewScheduler(opt *Option, ts Tasks, r *raft.Raft) *Scheduler {
	return &Scheduler{
		opt:         opt,
		Tasks:       ts,
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
					if err := t.Exec(context.TODO()); err != nil {
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
	if s.Tasks.Len() < 1 {
		return
	}

	now := time.Now()
	top := s.Tasks.Top().(*Task)
	if top.PlanExecTime.Before(now) || top.PlanExecTime.Equal(now) {
		wet := *top
		s.waitExec <- &wet

		top.PlanExecTime = top.CronExpression.Next(now)
		heap.Fix(s.Tasks, 0)
	}

	tick := s.opt.TaskRenewTick
	top = s.Tasks.Top().(*Task)
	if top.PlanExecTime.Sub(now) < tick {
		tick = top.PlanExecTime.Sub(now)
	}
	s.renewTick = time.NewTicker(tick)
}
