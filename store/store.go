package store

import (
	"cron-s/task"
	"time"
)

type Store struct {
	Task      *task.Task
	NodeId    string
	Ip        string
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func NewStore(t *task.Task) *Store {
	return &Store{
		Task: t,
	}
}
