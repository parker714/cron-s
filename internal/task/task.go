package task

import (
	ctx "context"
	"github.com/gorhill/cronexpr"
	"os/exec"
	"time"
)

const (
	StatusAdd = 1
	StatusDel = 2
)

// {"name":"task1", "cmd":"echo hello;", "cron_line":"*/5 * * * * * *"}`
type Task struct {
	Name           string               `json:"name"`
	Cmd            string               `json:"cmd"`
	CronLine       string               `json:"cron_line"`
	RunTime        time.Time            `json:"-"`
	CronExpression *cronexpr.Expression `json:"-"`
	Status         int                  `json:"-"`
}

func New() *Task {
	return &Task{}
}

func (t *Task) Exec() ([]byte, error) {
	cmd := exec.CommandContext(ctx.TODO(), "/bin/bash", "-c", t.Cmd)
	return cmd.CombinedOutput()
}
