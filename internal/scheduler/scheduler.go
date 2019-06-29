package scheduler

import (
	"cron-s/internal/conf"
	"cron-s/internal/routers"
	"cron-s/internal/tick"
	"cron-s/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/judwhite/go-svc/svc"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type scheduler struct {
	waitGroup util.WaitGroupWrapper

	raft   *raft.Raft
	engine *gin.Engine

	ts *tick.Tick
}

func New() *scheduler {
	return &scheduler{}
}

func (s *scheduler) Init(env svc.Environment) error {
	var err error
	if s.raft, err = s.newRaft(); err != nil {
		log.Error("schedule: newRaft err", err)
	}

	s.ts = tick.New(s.raft)

	r := routers.New(s.raft, s.ts)
	s.engine = gin.Default()
	s.engine.GET("/api/tasks", r.Tasks)
	s.engine.GET("/api/task/save", r.TaskSave)
	s.engine.GET("/api/task/del", r.TaskDel)
	s.engine.GET("/api/join", r.Join)

	return nil
}

func (s *scheduler) Start() error {
	log.Info("service start...")

	s.waitGroup.Wrap(func() {
		s.ts.Run()
	})

	if conf.Join != "" {
		err := s.joinCluster()
		if err != nil {
			log.Error("schedule: joinCluster err", err)
		}
	}

	if err := s.engine.Run(conf.HttpPort); err != nil {
		log.Error("schedule: listen http err", err)
	}
	return nil
}

func (s *scheduler) Stop() error {
	log.Info("service stop...")

	s.waitGroup.Wait()
	s.raft.Shutdown()

	return nil
}

func (s *scheduler) joinCluster() error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", conf.Join, conf.NodeId, conf.Bind)

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
