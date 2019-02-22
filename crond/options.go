package crond

import (
	"cron-s/internal/lg"
	"time"
)

type Options struct {
	LogLevel lg.LogLevel

	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
}

func NewOptions() *Options {
	return &Options{
		LogLevel: lg.INFO,

		EtcdEndpoints:   []string{"127.0.0.1:2379"},
		EtcdDialTimeout: 3 * time.Second,
	}
}
