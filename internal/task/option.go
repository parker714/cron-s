package task

import "time"

type Option struct {
	TaskRenewTick time.Duration
}

func NewOption() *Option {
	return &Option{
		TaskRenewTick: time.Second * 3,
	}
}
