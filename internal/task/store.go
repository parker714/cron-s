package task

import (
	"time"
)

type storage struct {
	Task      *Task
	NodeId    string
	Ip        string
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func NewStorage(t *Task) *storage {
	return &storage{
		Task: t,
	}
}

func (s *storage) Save() {
}
