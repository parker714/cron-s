package cronadmin

import (
	"cron-s/internal/lg"
	"time"
)

type Options struct {
	OutLogLevel lg.LogLevel

	HttpAddr         string
	HttpReadTimeout  time.Duration
	HttpWriteTimeout time.Duration

	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
}

func NewOptions() *Options {
	return &Options{
		OutLogLevel: lg.INFO,

		HttpAddr:         ":7570",
		HttpReadTimeout:  3 * time.Second,
		HttpWriteTimeout: 5 * time.Second,

		EtcdEndpoints:   []string{"127.0.0.1:2379"},
		EtcdDialTimeout: 3 * time.Second,
	}
}
