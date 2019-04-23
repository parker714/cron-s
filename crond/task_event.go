package crond

type taskEvent struct {
	Task *task `json:"task"`
	Cmd  int   `json:"cmd"` // 1:add, 2:del, 3:put, 4:get
}

const (
	TASK_ADD = 1
	TASK_DEL = 2
)

func NewTaskEvent(t *task, c int) *taskEvent {
	return &taskEvent{
		Task: t,
		Cmd:  c,
	}
}
