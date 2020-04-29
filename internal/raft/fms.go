package raft

import (
	"container/heap"
	"encoding/json"
	"github.com/parker714/cron-s/internal/task"
	"io"

	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type fms struct {
	tasks task.Tasks
}

func newFms(ts task.Tasks) *fms {
	return &fms{tasks: ts}
}

func (f *fms) Apply(l *raft.Log) interface{} {
	log.Debug("fms: Apply")

	t := task.New()
	if err := json.Unmarshal(l.Data, t); err != nil {
		log.Error("fms: Apply Unmarshal err", err)
		return nil
	}

	switch t.Status {
	case task.StatusAdd:
		f.tasks.Push(t)
	case task.StatusDel:
		f.tasks.Remove(t.Name)
	}

	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	log.Debug("fms: Snapshot")

	return newFmsSnapshot(f.tasks), nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	log.Debug("fpm: Restore")

	nts := task.NewTasks()
	if err := json.NewDecoder(serialized).Decode(nts); err != nil {
		return err
	}
	heap.Init(f.tasks)

	return nil
}
