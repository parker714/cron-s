package raft

import (
	"github.com/parker714/cron-s/internal/task"
	"io"
	"net"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftBoltdb "github.com/hashicorp/raft-boltdb"
)

// New app raft
func New(opt *Option, ts task.Tasks, lg io.Writer) (*raft.Raft, error) {
	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(opt.NodeID)

	LogStore, err := raftBoltdb.NewBoltStore(filepath.Join(opt.DataDir, "raft_log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftBoltdb.NewBoltStore(filepath.Join(opt.DataDir, "raft_stable.bolt"))
	if err != nil {
		return nil, err
	}

	snaps, err := raft.NewFileSnapshotStore(opt.DataDir, 1, lg)
	if err != nil {
		return nil, err
	}

	address, err := net.ResolveTCPAddr("tcp", opt.Bind)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, lg)
	if err != nil {
		return nil, err
	}

	fsm := newFms(ts)
	r, err := raft.NewRaft(rc, fsm, LogStore, stableStore, snaps, transport)
	if err != nil {
		return nil, err
	}

	if opt.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      rc.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(configuration)
	}

	return r, err
}
