package task

const (
	ADD = 1
	DEL = 2
)

type Event struct {
	Task *Task `json:"task"`
	Cmd  int   `json:"cmd"`
}

func NewEvent(t *Task, c int) *Event {
	return &Event{
		Task: t,
		Cmd:  c,
	}
}
