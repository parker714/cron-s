package job

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"time"
)

const (
	PUT = 0
	DEL = 1
)

// {"name":"job1","cmd":"echo hello;","cron_line":"*/5 * * * * * *"}`)
type Job struct {
	Name           string `json:"name"`
	Cmd            string `json:"cmd"`
	CronLine       string `json:"cron_line"`
	NextRunTime    time.Time
	CronExpression *cronexpr.Expression
}

type ChangeEvent struct {
	Key  string
	Type int8
	Job  *Job
}

type CompleteEvent struct {
	Key       string
	Job       *Job
	StartTime time.Time
	EndTime   time.Time
	Result    []byte
	Err       error
}

func Unmarshal(job []byte) (*Job, error) {
	j := &Job{}
	err := json.Unmarshal(job, j)
	if err != nil {
		return nil, err
	}

	j.CronExpression, err = cronexpr.Parse(j.CronLine)
	if err != nil {
		return nil, err
	}
	j.NextRunTime = j.CronExpression.Next(time.Now())

	return j, err
}
