package cronadmin

import (
	"cron-s/internal/lg"
	"cron-s/internal/server"
	"net/http"
)

type CronAdmin struct {
	Opts *Options

	lg *lg.Lg

	httpServer *http.Server

	server server.Server
}

func New(opts *Options) *CronAdmin {
	ca := &CronAdmin{
		Opts: opts,
	}

	ca.lg = lg.New("[cronadmin]", opts.OutLogLevel)

	return ca
}

func (ca *CronAdmin) Main() {
	var err error
	ca.lg.Logf(lg.INFO, "Main...")

	ca.InitHttpServer()
	ca.lg.Logf(lg.INFO, "Init Http Server Success... [%s]", ca.Opts.HttpAddr)

	ca.server, err = server.NewEtcd(ca.Opts.EtcdEndpoints)
	if err != nil {
		ca.lg.Logf(lg.ERROR, "NewEtcd err:%s, EtcdEndpoints:%s", err, ca.Opts.EtcdEndpoints)
		return
	}

	if err := ca.httpServer.ListenAndServe(); err != nil {
		ca.lg.Logf(lg.ERROR, "Http server err: %s", err)
		return
	}
}

func (ca *CronAdmin) Exit() {
	if ca.httpServer != nil {
		ca.httpServer.Close()
	}
	if ca.server != nil {
		ca.server.Close()
	}

	ca.lg.Logf(lg.INFO, "Exit...")
}
