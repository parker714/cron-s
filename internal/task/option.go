package task

import "time"

// Option struct
type Option struct {
	TaskRenewTick time.Duration
}

// NewOption returns option instance
func NewOption() *Option {
	return &Option{
		TaskRenewTick: time.Second * 3,
	}
}
