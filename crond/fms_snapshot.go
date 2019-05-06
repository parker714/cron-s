package crond

import (
	"encoding/json"
	"github.com/hashicorp/raft"
)

type fmsSnapshot struct {
	ctx *context
}

func (fs *fmsSnapshot) Persist(sink raft.SnapshotSink) error {
	fs.ctx.crond.log.Println("[DEBUG] fmsSnapshot: Persist")

	snapshotBytes, err := json.Marshal(fs.ctx.crond.tm.All())
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
	fs.ctx.crond.log.Println("[DEBUG] fmsSnapshot: Release")
}
