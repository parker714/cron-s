package crond

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io/ioutil"
	"net/http"
	"time"
)

func (c *Crond) initHttpServer() {
	mux := http.NewServeMux()

	//handleStatic := http.FileServer(http.Dir("../../cronadmin/static"))
	//mux.Handle("/", http.StripPrefix("/", handleStatic))

	mux.HandleFunc("/api/tasks", c.handleTasks)
	mux.HandleFunc("/api/task/save", c.handleTaskSave)
	mux.HandleFunc("/api/task/del", c.handleTaskDel)
	mux.HandleFunc("/api/join", c.handleJoin)

	c.httpServer = &http.Server{
		Addr:    c.opts.httpPort,
		Handler: mux,
	}
}

func (c *Crond) handleTasks(w http.ResponseWriter, r *http.Request) {
	// TODO
	c.mu.Lock()
	c.httpResponse(c.taskHeap, w)
	c.mu.Unlock()
}

func (c *Crond) handleTaskSave(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.log.Println("[WARN] http.handleTaskSave ReadAll err", err)
		return
	}

	t, err := Unmarshal(body)
	if err != nil {
		c.log.Println("[WARN] http.handleTaskSave Unmarshal err", err)
		return
	}

	c.taskEvent <- NewTaskEvent(t, TASK_ADD)
	c.scheduleTask()

	c.httpResponse("ok", w)
}
func (c *Crond) handleTaskDel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	t := &task{}
	err = json.Unmarshal(body, t)
	if err != nil {
		return
	}

	c.taskEvent <- NewTaskEvent(t, TASK_DEL)
	c.httpResponse("ok", w)
}

func (c *Crond) handleJoin(w http.ResponseWriter, r *http.Request) {
	nodeId := r.FormValue("nodeId")
	peerAddress := r.FormValue("peerAddress")

	index := c.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(peerAddress), 0, 3*time.Second)
	if err := index.Error(); err != nil {
		fmt.Println(err)
	}

	c.httpResponse("ok", w)
}

func (c *Crond) httpResponse(v interface{}, w http.ResponseWriter) {
	resp, err := json.Marshal(v)
	if err != nil {
		c.log.Println("[WARN] http.httpResponse, json marshal err", err)
		return
	}

	if _, err := w.Write(resp); err != nil {
		c.log.Println("[WARN] http.httpResponse, write err ", err)
	}
}
