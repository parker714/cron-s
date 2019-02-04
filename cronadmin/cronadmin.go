package cronadmin

import (
	"fmt"
	"net/http"
)

type CronAdmin struct {
	Opts *Options

	HttpServer *http.Server
}

func New(opts *Options) *CronAdmin {
	return &CronAdmin{
		Opts: opts,
	}
}

func (ca *CronAdmin) Main() {
	ca.InitHttpServer()

	if err := ca.HttpServer.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}

func (ca *CronAdmin) Exit() {
}
