package crond

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/raft"
	"io"
	"time"
)

type fms struct {
	ctx Context
}

func (f *fms) Apply(l *raft.Log) interface{} {
	fmt.Println("fms Apply", l)

	te := &taskEvent{}
	if err := json.Unmarshal(l.Data, &te); err != nil {
		f.ctx.crond.log.Println("[WARN] Apply task event err")
		return nil
	}

	// todo ??
	switch te.Cmd {
	case TASK_ADD:
		var err error
		te.Task.cronExpression, err = cronexpr.Parse(te.Task.CronLine)
		if err != nil {
			f.ctx.crond.log.Println("[WARN] Apply task event Parse err", err)
			return nil
		}
		te.Task.runTime = te.Task.cronExpression.Next(time.Now())

		heap.Push(f.ctx.crond.taskHeap, te.Task)
	case TASK_DEL:
		for i, task := range *f.ctx.crond.taskHeap {
			if te.Task.Name == task.Name {
				heap.Remove(f.ctx.crond.taskHeap, i)
			}
		}
	}
	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Println("fms Snapshot")

	return &fmsSnapshot{
		ctx: Context{crond: f.ctx.crond},
	}, nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	fmt.Println("fms Restore")

	th := &taskHeap{}
	if err := json.NewDecoder(serialized).Decode(th); err != nil {
		return err
	}

	// todo ??
	for _, t := range *th {
		f.ctx.crond.taskHeap.Push(t)
	}

	return nil
}
