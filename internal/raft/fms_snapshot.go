package raft

import (
	"cron-s/internal/task"
	"encoding/json"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type fmsSnapshot struct {
	taskData *task.Data
}

func newFmsSnapshot(td *task.Data) *fmsSnapshot {
	return &fmsSnapshot{taskData: td}
}

func (fs *fmsSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("fmsSnapshot: Persist")

	snapshotBytes, err := json.Marshal(fs.taskData.All())
	if err != nil {
		return sink.Cancel()
	}
	if _, err := sink.Write(snapshotBytes); err != nil {
		return sink.Cancel()
	}
	if err := sink.Close(); err != nil {
		return sink.Cancel()
	}

	return nil
}

func (fs *fmsSnapshot) Release() {
	log.Debug("fmsSnapshot: Release")
}
