package crond

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftBoltdb "github.com/hashicorp/raft-boltdb"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (c *Crond) newRaft(opts *options) (*raft.Raft, error) {
	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(opts.nodeId)

	LogStore, err := raftBoltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-Log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftBoltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	snaps, err := raft.NewFileSnapshotStore(opts.dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	address, err := net.ResolveTCPAddr("tcp", opts.bind)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	fsm := &Fms{
		ctx: &Context{Crond: c},
	}

	r, err := raft.NewRaft(rc, fsm, LogStore, stableStore, snaps, transport)
	if err != nil {
		return nil, err
	}

	if opts.bootstrap {
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

func (c *Crond) joinCluster(opts *options) error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", opts.join, opts.nodeId, opts.bind)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("addCluster url %s err %s", url, err)
	}

	return nil
}
