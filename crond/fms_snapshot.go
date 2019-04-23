package crond

import (
	"encoding/json"
	"github.com/hashicorp/raft"
)

type FmsSnapshot struct {
	ctx *Context
}

func (fs *FmsSnapshot) Persist(sink raft.SnapshotSink) error {
	fs.ctx.Crond.Log.Println("[DEBUG] fmsSnapshot: Persist")

	snapshotBytes, err := json.Marshal(fs.ctx.Crond.TaskHeap)
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

func (fs *FmsSnapshot) Release() {
	fs.ctx.Crond.Log.Println("[DEBUG] fmsSnapshot: Release")
}
