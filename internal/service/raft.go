package service

import (
	"cron-s/internal/conf"
	"cron-s/internal/fms"
	"fmt"
	"github.com/hashicorp/raft"
	raftBoltdb "github.com/hashicorp/raft-boltdb"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (s *service) newRaft() (*raft.Raft, error) {
	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(conf.NodeId)

	LogStore, err := raftBoltdb.NewBoltStore(filepath.Join(conf.DataDir, "raft-Log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftBoltdb.NewBoltStore(filepath.Join(conf.DataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	snaps, err := raft.NewFileSnapshotStore(conf.DataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	address, err := net.ResolveTCPAddr("tcp", conf.Bind)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	fsm := fms.New()
	r, err := raft.NewRaft(rc, fsm, LogStore, stableStore, snaps, transport)
	if err != nil {
		return nil, err
	}

	if conf.Bootstrap {
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

func (s *service) joinCluster() error {
	url := fmt.Sprintf("http://%s/api/join?nodeId=%s&peerAddress=%s", conf.Join, conf.NodeId, conf.Bind)

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
