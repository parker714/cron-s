package scheduler

import (
	raft2 "cron-s/internal/raft"
	"cron-s/internal/task"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/judwhite/go-svc/svc"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type scheduler struct {
	opt      *Option
	taskData *task.Data
	engine   *gin.Engine

	raft          *raft.Raft
	taskScheduler *task.Scheduler
}

func New(opt *Option) *scheduler {
	gin.SetMode(gin.ReleaseMode)
	return &scheduler{
		opt:      opt,
		taskData: task.NewData(),
		engine:   gin.Default(),
	}
}

func (s *scheduler) Init(env svc.Environment) (err error) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	lev, err := log.ParseLevel(s.opt.LogLevel)
	if err != nil {
		return
	}
	log.SetLevel(lev)

	if s.opt.Join != "" {
		s.opt.Raft.Bootstrap = false
	}
	if s.raft, err = raft2.New(s.opt.Raft, s.taskData, log.StandardLogger().Out); err != nil {
		log.Error("App scheduler newRaft err,", err)
	}

	s.taskScheduler = task.NewScheduler(s.opt.Task, s.taskData, s.raft)

	r := newRouter(s.taskScheduler)
	s.engine.GET("/api/tasks", r.Tasks)
	s.engine.GET("/api/task/save", r.TaskSave)
	s.engine.GET("/api/task/del", r.TaskDel)
	s.engine.GET("/api/join", r.Join)

	log.Info("App scheduler init")
	return nil
}

func (s *scheduler) Start() error {
	log.Debug("App scheduler start")

	s.taskScheduler.Run()

	if s.opt.Join != "" {
		err := s.joinCluster()
		if err != nil {
			log.Warn("App scheduler joinCluster err,", err)
		}
	}

	if err := s.engine.Run(s.opt.HttpPort); err != nil {
		log.Error("App scheduler listen http err,", err)
	}
	return nil
}

func (s *scheduler) Stop() error {
	log.Info("App scheduler stop")

	s.raft.Shutdown()
	s.taskScheduler.WaitGroup.Wait()

	return nil
}

func (s *scheduler) joinCluster() error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", s.opt.Join, s.opt.Raft.NodeId, s.opt.Raft.Bind)

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
