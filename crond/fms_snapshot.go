package crond

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
)

type fmsSnapshot struct {
	ctx Context
}

func (f *fmsSnapshot) Persist(sink raft.SnapshotSink) error {
	snapshotBytes, err := json.Marshal(f.ctx.crond.taskHeap)
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

func (f *fmsSnapshot) Release() {
	fmt.Println("fmsSnapshot Release")
}
