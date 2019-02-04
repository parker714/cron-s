package server

import "cron-s/internal/job"

const (
	JobsKey = "/cron/jobs/"
)

type Server interface {
	Get() (map[string]*job.Job, error)
	Watch(chan *job.ChangeEvent)
	Lock(key string, a func()) error
}
