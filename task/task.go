package task

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"time"
)

type Task struct {
	Name           string               `json:"name"`
	Cmd            string               `json:"cmd"`
	CronLine       string               `json:"cron_line"`
	RunTime        time.Time            `json:"-"`
	CronExpression *cronexpr.Expression `json:"-"`
}

// task protocol
// {"name":"task1","cmd":"echo hello;","cron_line":"*/5 * * * * * *"}`

func Marshal(t *Task) (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}
