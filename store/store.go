package store

import (
	"crond/tasks"
	"time"
)

type Store struct {
	Task      *tasks.Task
	NodeId    string
	Ip        string
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func NewStore(t *tasks.Task) *Store {
	return &Store{
		Task: t,
	}
}

func (s *Store) Save() {
}
