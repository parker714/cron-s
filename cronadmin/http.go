package cronadmin

import (
	"cron-s/internal/lg"
	"cron-s/internal/server"
	"cron-s/internal/task"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (ca *CronAdmin) handleTasks(w http.ResponseWriter, r *http.Request) {
	tasksMap, err := ca.server.Get(server.JobsKey)
	if err != nil {
		return
	}

	tasks := make([]*task.Task, 0)
	for _, t := range tasksMap {
		tasks = append(tasks, t)
	}

	ca.httpResponse(tasks, w)
}
func (ca *CronAdmin) handleTaskSave(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	t := &task.Task{}
	err = json.Unmarshal(body, t)
	if err != nil {
		return
	}

	err = ca.server.Add(t)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := w.Write([]byte("ok")); err != nil {
		fmt.Println(err)
	}
}
func (ca *CronAdmin) handleTaskDel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	t := &task.Task{}
	err = json.Unmarshal(body, t)
	if err != nil {
		return
	}

	err = ca.server.Del(t)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := w.Write([]byte("ok")); err != nil {
		fmt.Println(err)
	}
}

func (ca *CronAdmin) httpResponse(v interface{}, w http.ResponseWriter) {
	resp, err := json.Marshal(v)
	if err != nil {
		ca.lg.Logf(lg.ERROR, "http response json marshal err: %s", err)
		return
	}

	if _, err := w.Write(resp); err != nil {
		ca.lg.Logf(lg.ERROR, "http response err: %s", err)
	}
}

func (ca *CronAdmin) InitHttpServer() {
	mux := http.NewServeMux()

	handleStatic := http.FileServer(http.Dir("../../cronadmin/static"))
	mux.Handle("/", http.StripPrefix("/", handleStatic))

	mux.HandleFunc("/api/tasks", ca.handleTasks)
	mux.HandleFunc("/api/task/save", ca.handleTaskSave)
	mux.HandleFunc("/api/task/del", ca.handleTaskDel)

	ca.httpServer = &http.Server{
		Addr:         ca.Opts.HttpAddr,
		ReadTimeout:  ca.Opts.HttpReadTimeout,
		WriteTimeout: ca.Opts.HttpWriteTimeout,
		Handler:      mux,
	}
}
