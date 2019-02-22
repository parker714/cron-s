package server

import "cron-s/internal/task"

const (
	JobsKey = "/cron/tasks/"
)

type Server interface {
	Add(task *task.Task) error
	Del(task *task.Task) error
	Get(key string) (map[string]*task.Task, error)
	Watch(me chan *task.ModifyEvent)
	Lock(key string, do func()) error
	Close() error
}
