package agent

import (
	"github.com/parker714/cron-s/internal/raft"
	"github.com/parker714/cron-s/internal/task"
)

// Option is app config struct
type Option struct {
	LogLevel string
	HTTPPort string
	Join     string
	Task     *task.Option
	Raft     *raft.Option
}

// NewOption returns app config instance
func NewOption() *Option {
	return &Option{
		LogLevel: "debug",
		Task:     task.NewOption(),
		Raft:     raft.NewOption(),
	}
}
