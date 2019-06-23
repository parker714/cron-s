package fms

import (
	"cron-s/internal/store"
	"encoding/json"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type fmsSnapshot struct {
}

func (fs *fmsSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("fmsSnapshot: Persist")

	snapshotBytes, err := json.Marshal(store.All())
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
