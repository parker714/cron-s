package tasks

import (
	"time"
)

type Storage struct {
	Task      *Task
	NodeId    string
	Ip        string
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func NewStorage(t *Task) *Storage {
	return &Storage{
		Task: t,
	}
}

func (s *Storage) Save() {
}
