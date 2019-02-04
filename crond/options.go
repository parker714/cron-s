package crond

import (
	"cron-s/internal/lg"
	"log"
	"time"
)

type Options struct {
	Logger    *log.Logger
	LogLevel  lg.LogLevel
	LogPrefix string

	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
}

func NewOptions() *Options {
	return &Options{
		LogLevel:  lg.INFO,
		LogPrefix: "[crond]",

		EtcdEndpoints:   []string{"127.0.0.1:2379"},
		EtcdDialTimeout: 3 * time.Second,
	}
}
