package raft

import (
	cheap "container/heap"
	"cron-s/internal/task"
	"encoding/json"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"io"
)

type fms struct {
	taskHeap task.Heap
}

func newFms(th task.Heap) *fms {
	return &fms{taskHeap: th}
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
		f.taskHeap.Push(t)
	case task.StatusDel:
		f.taskHeap.Remove(t)
	}

	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	log.Debug("fms: Snapshot")

	return newFmsSnapshot(f.taskHeap), nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	log.Debug("fpm: Restore")

	nh := task.NewHeap()
	if err := json.NewDecoder(serialized).Decode(nh); err != nil {
		return err
	}
	cheap.Init(f.taskHeap)

	return nil
}
