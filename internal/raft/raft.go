package raft

import (
	"cron-s/internal/task"
	"github.com/hashicorp/raft"
	raftBoltdb "github.com/hashicorp/raft-boltdb"
	"io"
	"net"
	"path/filepath"
	"time"
)

func New(opt *Option, td *task.Data, lg io.Writer) (*raft.Raft, error) {
	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(opt.NodeId)

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

	fsm := newFms(td)
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
