package tasks

import (
	"time"
)

type Store struct {
	Task      *Task
	NodeId    string
	Ip        string
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func NewStore(t *Task) *Store {
	return &Store{
		Task: t,
	}
}

func (s *Store) Save() {
}
