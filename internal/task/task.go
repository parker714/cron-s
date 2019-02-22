package task

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"time"
)

const (
	PUT = 0
	DEL = 1
)

type Task struct {
	Name           string               `json:"name"`
	Cmd            string               `json:"cmd"`
	CronLine       string               `json:"cron_line"`
	NextRunTime    time.Time            `json:"-"`
	CronExpression *cronexpr.Expression `json:"-"`
}

type ModifyEvent struct {
	Name string
	Task *Task
	Type int8
}

type Schedule struct {
	Name      string
	Task      *Task
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

// task protocol
// {"name":"job1","cmd":"echo hello;","cron_line":"*/5 * * * * * *"}`)
func Unmarshal(job []byte) (*Task, error) {
	t := &Task{}
	err := json.Unmarshal(job, t)
	if err != nil {
		return nil, err
	}

	t.CronExpression, err = cronexpr.Parse(t.CronLine)
	if err != nil {
		return nil, err
	}
	t.NextRunTime = t.CronExpression.Next(time.Now())

	return t, err
}

func Marshal(task *Task) (string, error) {
	bytes, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}
