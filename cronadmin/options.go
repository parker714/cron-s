package cronadmin

import "time"

type Options struct {
	HttpAddr         string
	HttpReadTimeout  time.Duration
	HttpWriteTimeout time.Duration
}

func NewOptions() *Options {
	return &Options{
		HttpAddr:         ":7570",
		HttpReadTimeout:  3 * time.Second,
		HttpWriteTimeout: 5 * time.Second,
	}
}
