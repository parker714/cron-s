package crond

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"time"
)

type task struct {
	Name           string               `json:"name"`
	Cmd            string               `json:"cmd"`
	CronLine       string               `json:"cron_line"`
	runTime        time.Time            `json:"-"`
	cronExpression *cronexpr.Expression `json:"-"`
}

// task protocol
// {"name":"task1","cmd":"echo hello;","cron_line":"*/5 * * * * * *"}`
func Unmarshal(tb []byte) (*task, error) {
	t := &task{}
	err := json.Unmarshal(tb, t)
	if err != nil {
		return nil, err
	}

	t.cronExpression, err = cronexpr.Parse(t.CronLine)
	if err != nil {
		return nil, err
	}
	t.runTime = t.cronExpression.Next(time.Now())

	return t, err
}

func Marshal(t *task) (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}
