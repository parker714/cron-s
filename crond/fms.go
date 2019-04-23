package crond

import (
	"container/heap"
	"cron-s/task"
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
)

type Fms struct {
	ctx *Context
}

func (f *Fms) Apply(l *raft.Log) interface{} {
	f.ctx.Crond.Log.Println("[DEBUG] fms: Apply")

	te := &task.Event{}
	if err := json.Unmarshal(l.Data, te); err != nil {
		f.ctx.Crond.Log.Println("[WARN] fms: Apply Unmarshal err", err)
		return nil
	}

	if err := f.ctx.Crond.HandleTaskEvent(te); err != nil {
		f.ctx.Crond.Log.Println("[WARN] fms: Apply HandleTaskEvent err", err)
		return nil
	}

	return nil
}

func (f *Fms) Snapshot() (raft.FSMSnapshot, error) {
	f.ctx.Crond.Log.Println("[DEBUG] fms: Snapshot")

	return &FmsSnapshot{
		ctx: &Context{Crond: f.ctx.Crond},
	}, nil
}

func (f *Fms) Restore(serialized io.ReadCloser) error {
	f.ctx.Crond.Log.Println("[DEBUG] fpm: Restore")

	if err := json.NewDecoder(serialized).Decode(f.ctx.Crond.TaskHeap); err != nil {
		return err
	}
	heap.Init(f.ctx.Crond.TaskHeap)

	return nil
}
