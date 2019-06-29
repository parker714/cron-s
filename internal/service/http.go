package service

import (
	"cron-s/internal/conf"
	"cron-s/internal/store"
	"cron-s/pkg/tasks"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func (s *service) initMux() {
	mux := http.NewServeMux()

	handleStatic := http.FileServer(http.Dir("web"))
	mux.Handle("/", http.StripPrefix("/", handleStatic))

	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/task/save", s.handleTaskSave)
	mux.HandleFunc("/api/task/del", s.handleTaskDel)
	mux.HandleFunc("/api/join", s.handleJoin)

	s.httpServer = &http.Server{
		Addr:    conf.HttpPort,
		Handler: mux,
	}
}

func (s *service) handleTasks(w http.ResponseWriter, r *http.Request) {
	s.httpResponse(store.All(), w)
}

func (s *service) handleTaskSave(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	t := tasks.NewTask()
	err = json.Unmarshal(body, t)
	if err != nil {
		log.Error("schedule: http.handleTaskSave Unmarshal err", err)
		return
	}
	t.CronExpression, err = cronexpr.Parse(t.CronLine)
	if err != nil {
		log.Error("schedule: http.handleTaskSave Parse err", err)
		return
	}
	t.RunTime = t.CronExpression.Next(time.Now())

	store.Add(t)
	s.scheduleTask()

	s.httpResponse("ok", w)
}

func (s *service) handleTaskDel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	t := tasks.NewTask()
	err = json.Unmarshal(body, t)
	if err != nil {
		log.Error("schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	store.Del(t)
	s.scheduleTask()

	s.httpResponse("ok", w)
}

func (s *service) handleJoin(w http.ResponseWriter, r *http.Request) {
	nodeId := r.FormValue("nodeId")
	peerAddress := r.FormValue("peerAddress")

	index := s.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(peerAddress), 0, 3*time.Second)
	if err := index.Error(); err != nil {
		log.Error("schedule: http.handleJoin err", err)
		return
	}

	s.httpResponse("ok", w)
}

func (s *service) httpResponse(v interface{}, w http.ResponseWriter) {
	resp, err := json.Marshal(v)
	if err != nil {
		log.Error("schedule: http.httpResponse, Marshal err", err)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Error("schedule: http.httpResponse, Write err ", err)
		return
	}
}
