package crond

import (
	"cron-s/tasks"
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
)

type fms struct {
	ctx *context
}

func (f *fms) Apply(l *raft.Log) interface{} {
	f.ctx.crond.log.Println("[DEBUG] fms: Apply")

	t := tasks.NewTask()
	if err := json.Unmarshal(l.Data, t); err != nil {
		f.ctx.crond.log.Println("[WARN] fms: Apply Unmarshal err", err)
		return nil
	}

	switch t.Status {
	case tasks.StatusAdd:
		f.ctx.crond.tm.Add(t)
	case tasks.StatusDel:
		f.ctx.crond.tm.Del(t)
	}

	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	f.ctx.crond.log.Println("[DEBUG] fms: Snapshot")

	return &fmsSnapshot{
		ctx: &context{crond: f.ctx.crond},
	}, nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	f.ctx.crond.log.Println("[DEBUG] fpm: Restore")

	nh := tasks.NewHeap()
	if err := json.NewDecoder(serialized).Decode(nh); err != nil {
		return err
	}
	f.ctx.crond.tm.Init(nh)

	return nil
}
