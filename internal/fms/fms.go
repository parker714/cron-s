package fms

import (
	"cron-s/internal/store"
	"cron-s/pkg/tasks"
	"encoding/json"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"io"
)

type fms struct {
}

func New() *fms {
	return &fms{}
}

func (f *fms) Apply(l *raft.Log) interface{} {
	log.Debug("fms: Apply")

	t := tasks.NewTask()
	if err := json.Unmarshal(l.Data, t); err != nil {
		log.Error("fms: Apply Unmarshal err", err)
		return nil
	}

	switch t.Status {
	case tasks.StatusAdd:
		store.Add(t)
	case tasks.StatusDel:
		store.Del(t)
	}

	return nil
}

func (f *fms) Snapshot() (raft.FSMSnapshot, error) {
	log.Debug("fms: Snapshot")

	return &fmsSnapshot{}, nil
}

func (f *fms) Restore(serialized io.ReadCloser) error {
	log.Debug("fpm: Restore")

	nh := tasks.NewHeap()
	if err := json.NewDecoder(serialized).Decode(nh); err != nil {
		return err
	}
	store.Init(nh)

	return nil
}
