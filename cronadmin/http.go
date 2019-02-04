package cronadmin

import (
	"fmt"
	"net/http"
)

func (ca *CronAdmin) handleCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	if _, err := w.Write([]byte("create err")); err != nil {
		fmt.Println(err)
	}
}
func (ca *CronAdmin) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("update")); err != nil {
		fmt.Println(err)
	}
}
func (ca *CronAdmin) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("retrieve err")); err != nil {
		fmt.Println(err)
	}
}
func (ca *CronAdmin) handleDelete(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("delete err")); err != nil {
		fmt.Println(err)
	}
}

func (ca *CronAdmin) InitHttpServer() {
	mux := http.NewServeMux()

	handleStatic := http.FileServer(http.Dir("../../cronadmin/static"))
	mux.Handle("/", http.StripPrefix("/", handleStatic))

	mux.HandleFunc("/api/job/create", ca.handleCreate)
	mux.HandleFunc("/api/job/update", ca.handleUpdate)
	mux.HandleFunc("/api/job/retrieve", ca.handleRetrieve)
	mux.HandleFunc("/api/job/delete", ca.handleDelete)

	ca.HttpServer = &http.Server{
		Addr:         ca.Opts.HttpAddr,
		ReadTimeout:  ca.Opts.HttpReadTimeout,
		WriteTimeout: ca.Opts.HttpWriteTimeout,
		Handler:      mux,
	}
}
