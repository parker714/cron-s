package service

import (
	"cron-s/internal/conf"
	"cron-s/internal/store"
	"cron-s/pkg/tasks"
	"cron-s/pkg/util"
	"github.com/hashicorp/raft"
	"github.com/judwhite/go-svc/svc"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type service struct {
	waitGroup util.WaitGroupWrapper

	scheduleTaskTick <-chan time.Time
	waitExecTask     chan *tasks.Task
	waitStoreTask    chan *tasks.Store

	raft       *raft.Raft
	httpServer *http.Server
}

func NewService() *service {
	return &service{
		scheduleTaskTick: time.Tick(conf.DefaultScheduleTaskTick),
		waitExecTask:     make(chan *tasks.Task, 100),
		waitStoreTask:    make(chan *tasks.Store, 100),
	}
}

func (s *service) Init(env svc.Environment) error {
	fm := &log.TextFormatter{}
	fm.DisableColors = false
	fm.FullTimestamp = true
	fm.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(fm)

	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)

	return nil
}

func (s *service) Start() error {
	log.Info("service start...")

	s.waitGroup.Wrap(func() {
		for {
			select {
			case <-s.scheduleTaskTick:
				s.scheduleTask()
			case t := <-s.waitExecTask:
				log.Debug("schedule: exec Task")

				s.waitGroup.Wrap(func() {
					if s.raft.State() != raft.Leader {
						return
					}

					ss := tasks.NewStore(t)
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
	})

	var err error
	if s.raft, err = s.newRaft(); err != nil {
		log.Error("schedule: newRaft err", err)
	}
	if conf.Join != "" {
		err := s.joinCluster()
		if err != nil {
			log.Error("schedule: joinCluster err", err)
		}
	}

	s.initMux()
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Error("schedule: listen http err", err)
	}
	return nil
}

func (s *service) Stop() error {
	log.Info("service stop...")

	s.waitGroup.Wait()
	s.raft.Shutdown()

	return nil
}

func (s *service) scheduleTask() {
	if store.Len() < 1 {
		return
	}

	now := time.Now()
	top := store.Top()
	if top.RunTime.Before(now) || top.RunTime.Equal(now) {
		wet := *top
		s.waitExecTask <- &wet

		top.RunTime = top.CronExpression.Next(now)
		store.Fix(0)
	}

	tick := conf.DefaultScheduleTaskTick
	if store.Top().RunTime.Sub(now) < tick {
		tick = store.Top().RunTime.Sub(now)
	}
	s.scheduleTaskTick = time.Tick(tick)
}
