package scheduler

import (
	"cron-s/internal/conf"
	raft2 "cron-s/internal/raft"
	"cron-s/internal/routers"
	"cron-s/internal/task"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/judwhite/go-svc/svc"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type scheduler struct {
	cf       *conf.Config
	taskData *task.Data
	engine   *gin.Engine

	raft          *raft.Raft
	taskScheduler *task.Scheduler
}

func New(cf *conf.Config) *scheduler {
	return &scheduler{
		cf:       cf,
		taskData: task.NewData(),
		engine:   gin.Default(),
	}
}

func (s *scheduler) Init(env svc.Environment) (err error) {
	log.Debug("App scheduler init")

	if s.raft, err = raft2.New(s.cf.Raft, s.taskData); err != nil {
		log.Error("App scheduler newRaft err,", err)
	}

	s.taskScheduler = task.NewScheduler(s.cf, s.taskData, s.raft)

	r := routers.New(s.taskScheduler)
	s.engine.GET("/api/tasks", r.Tasks)
	s.engine.GET("/api/task/save", r.TaskSave)
	s.engine.GET("/api/task/del", r.TaskDel)
	s.engine.GET("/api/join", r.Join)

	return nil
}

func (s *scheduler) Start() error {
	log.Debug("App scheduler start")

	s.taskScheduler.Run()

	if s.cf.Join != "" {
		err := s.joinCluster()
		if err != nil {
			log.Warn("App scheduler joinCluster err,", err)
		}
	}

	if err := s.engine.Run(s.cf.HttpPort); err != nil {
		log.Error("App scheduler listen http err,", err)
	}
	return nil
}

func (s *scheduler) Stop() error {
	log.Debug("App scheduler stop")

	s.raft.Shutdown()
	s.taskScheduler.WaitGroup.Wait()

	return nil
}

func (s *scheduler) joinCluster() error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", s.cf.Join, s.cf.Raft.NodeId, s.cf.Raft.Bind)

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
