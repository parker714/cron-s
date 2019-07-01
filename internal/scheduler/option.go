package scheduler

import (
	"cron-s/internal/raft"
	"cron-s/internal/task"
)

type Option struct {
	LogLevel string
	HttpPort string
	Join     string
	Task     *task.Option
	Raft     *raft.Option
}

func NewOption() *Option {
	return &Option{
		LogLevel: "info",
		Task:     task.NewOption(),
		Raft:     raft.NewOption(),
	}
}
