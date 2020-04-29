package raft

import (
	"encoding/json"
	"github.com/parker714/cron-s/internal/task"

	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type fmsSnapshot struct {
	tasks task.Tasks
}

func newFmsSnapshot(th task.Tasks) *fmsSnapshot {
	return &fmsSnapshot{tasks: th}
}

func (fs *fmsSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("fmsSnapshot: Persist")

	snapshotBytes, err := json.Marshal(fs.tasks)
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
