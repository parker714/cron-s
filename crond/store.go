package crond

import "time"

type store struct {
	task      *task
	startTime time.Time
	endTime   time.Time
	result    []byte
	err       error
}

func NewStore() *store {
	return &store{}
}
