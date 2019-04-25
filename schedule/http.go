package schedule

import (
	"crond/tasks"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/raft"
	"io/ioutil"
	"net/http"
	"time"
)

func (c *Schedule) initHttpServer() {
	mux := http.NewServeMux()

	handleStatic := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.StripPrefix("/", handleStatic))

	mux.HandleFunc("/api/tasks", c.handleTasks)
	mux.HandleFunc("/api/task/save", c.handleTaskSave)
	mux.HandleFunc("/api/task/del", c.handleTaskDel)
	mux.HandleFunc("/api/join", c.handleJoin)

	c.httpServer = &http.Server{
		Addr:    c.opts.httpPort,
		Handler: mux,
	}
}

func (c *Schedule) handleTasks(w http.ResponseWriter, r *http.Request) {
	c.httpResponse(tasks.All(), w)
}

func (c *Schedule) handleTaskSave(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	t := tasks.NewTask()
	err = json.Unmarshal(body, t)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.handleTaskSave Unmarshal err", err)
		return
	}
	t.CronExpression, err = cronexpr.Parse(t.CronLine)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.handleTaskSave Parse err", err)
		return
	}
	t.RunTime = t.CronExpression.Next(time.Now())

	tasks.Add(t)
	c.scheduleTask()

	c.httpResponse("ok", w)
}

func (c *Schedule) handleTaskDel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	t := tasks.NewTask()
	err = json.Unmarshal(body, t)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	tasks.Del(t)
	c.scheduleTask()

	c.httpResponse("ok", w)
}

func (c *Schedule) handleJoin(w http.ResponseWriter, r *http.Request) {
	nodeId := r.FormValue("nodeId")
	peerAddress := r.FormValue("peerAddress")

	index := c.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(peerAddress), 0, 3*time.Second)
	if err := index.Error(); err != nil {
		c.Log.Println("[WARN] schedule: http.handleJoin err", err)
		return
	}

	c.httpResponse("ok", w)
}

func (c *Schedule) httpResponse(v interface{}, w http.ResponseWriter) {
	resp, err := json.Marshal(v)
	if err != nil {
		c.Log.Println("[WARN] schedule: http.httpResponse, Marshal err", err)
		return
	}

	if _, err := w.Write(resp); err != nil {
		c.Log.Println("[WARN] schedule: http.httpResponse, Write err ", err)
		return
	}
}
