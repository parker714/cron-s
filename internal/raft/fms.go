package raft

import (
	"cron-s/internal/task"
	"encoding/json"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"io"
)

type fms struct {
	taskData *task.Data
}

func newFms(td *task.Data) *fms {
	return &fms{taskData: td}
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
		f.taskData.Add(t)
	case task.StatusDel:
		f.taskData.Del(t)
	}

	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	log.Debug("fms: Snapshot")

	return newFmsSnapshot(f.taskData), nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	log.Debug("fpm: Restore")

	nh := task.NewHeap()
	if err := json.NewDecoder(serialized).Decode(nh); err != nil {
		return err
	}
	f.taskData.Init(nh)

	return nil
}
