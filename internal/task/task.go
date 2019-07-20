package task

import (
	ctx "context"
	"os/exec"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	// StatusAdd add task flag
	StatusAdd = 1
	// StatusDel del task flag
	StatusDel = 2
)

// Task struct
// {"name":"task1", "cmd":"echo hello;", "cron_line":"*/5 * * * * * *"}`
type Task struct {
	Name            string               `json:"name"`
	Cmd             string               `json:"cmd"`
	CronLine        string               `json:"cron_line"`
	PlanExecTime    time.Time            `json:"-"`
	CronExpression  *cronexpr.Expression `json:"-"`
	Status          int                  `json:"-"`
	NodeID          string               `json:"-"`
	IP              string               `json:"-"`
	ActualStartTime time.Time            `json:"-"`
	ActualEndTime   time.Time            `json:"-"`
	Result          []byte               `json:"-"`
}

// New returns task
func New() *Task {
	return &Task{}
}

// Exec task
func (t *Task) Exec(ctx ctx.Context) (err error) {
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", t.Cmd)
	t.Result, err = cmd.CombinedOutput()
	return
}

// Save task
func (t *Task) Save() {
}
