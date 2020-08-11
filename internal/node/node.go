package node

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	raftBoltDB "github.com/hashicorp/raft-boltdb"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
	"up/project/raft/store"
)

type Node interface {
	Run() error
}

type node struct {
	opt *option

	isLeader     int32
	stableStore  *raftBoltDB.BoltStore
	raft         *raft.Raft
	raftNotifyCh <-chan bool

	store store.Store
}

func New(opt *option) Node {
	return &node{
		opt: opt,
	}
}

func (n *node) setupRaft() error {
	defer func() {
		if n.raft == nil && n.stableStore != nil {
			if err := n.stableStore.Close(); err != nil {
				log.Printf("[ERR] node: setupRaft close Raft store err, %v", err)
			}
		}
	}()

	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(n.opt.nodeID)

	// test
	rc.LogLevel = "error"
	rc.SnapshotInterval = 20 * time.Second // 执行快照时间
	rc.SnapshotThreshold = 1               // 日志数目

	if err := os.MkdirAll(n.opt.dataDir, os.ModePerm); err != nil {
		return err
	}

	logStore, err := raftBoltDB.NewBoltStore(filepath.Join(n.opt.dataDir, "raft-log.bolt"))
	if err != nil {
		return err
	}

	n.stableStore, err = raftBoltDB.NewBoltStore(filepath.Join(n.opt.dataDir, "raft-stable.bolt"))
	if err != nil {
		return err
	}

	snaps, err := raft.NewFileSnapshotStore(n.opt.dataDir, 1, os.Stderr)
	if err != nil {
		return err
	}

	address, err := net.ResolveTCPAddr("tcp", n.opt.bind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 3*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	if n.opt.bootstrap {
		hasState, err := raft.HasExistingState(logStore, n.stableStore, snaps)
		if err != nil {
			return err
		}

		if !hasState {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      rc.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			}

			if err := raft.BootstrapCluster(rc, logStore, n.stableStore, snaps, transport, configuration); err != nil {
				return err
			}
		}
	}

	raftNotifyCh := make(chan bool, 1)
	rc.NotifyCh = raftNotifyCh
	n.raftNotifyCh = raftNotifyCh

	n.raft, err = raft.NewRaft(rc, n, logStore, n.stableStore, snaps, transport)
	return err
}

func (n *node) monitorLeadership() {
	for {
		select {
		case leader := <-n.raftNotifyCh:
			if leader {
				atomic.StoreInt32(&n.isLeader, 1)
			} else {
				atomic.StoreInt32(&n.isLeader, 0)
			}
		}
	}
}

func (n *node) joinCluster() error {
	url := fmt.Sprintf("http://%s/join?nodeId=%s&peerAddress=%s", n.opt.join, n.opt.nodeID, n.opt.bind)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("joinCluster fail, err: %s", err)
		}

		return fmt.Errorf("joinCluster fail, err: %s", body)
	}

	return nil
}

func (n *node) Run() error {
	if err := n.setupRaft(); err != nil {
		return err
	}

	go n.monitorLeadership()

	if n.opt.join != "" {
		if err := n.joinCluster(); err != nil {
			return err
		}
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/join", n.join)
	router.GET("/set", n.set)
	router.GET("/get", n.get)
	return router.Run(n.opt.httpPort)
}
