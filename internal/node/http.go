package node

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

func (n *node) join(c *gin.Context) {
	nodeId := c.Query("nodeId")
	peerAddress := c.Query("peerAddress")

	index := n.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(peerAddress), 0, 3*time.Second)
	if err := index.Error(); err != nil {
		log.Printf("[Error] node id:%s peerAddress: %s join cluster fail, err: %s\n", nodeId, peerAddress, err)
		c.String(http.StatusInternalServerError, "node join cluster fail")
		return
	}

	c.String(http.StatusOK, "ok")
}

func (n *node) set(c *gin.Context) {
	if v := atomic.LoadInt32(&n.isLeader); v != 1 {
		c.String(http.StatusInternalServerError, "node not leader, not allow set")
		return
	}

	key := c.Query("key")
	value := c.Query("value")

	logEntryBytes, err := json.Marshal(logEntryData{Key: key, Value: value})
	if err != nil {
		log.Printf("[Error] node marshal data fail, err:%v\n", err)
		c.String(http.StatusInternalServerError, "node marshal data fail")
		return
	}
	af := n.raft.Apply(logEntryBytes, 3*time.Second)
	if err := af.Error(); err != nil {
		log.Printf("[Error] node apply data fail, err:%v\n", err)
		c.String(http.StatusInternalServerError, "node apply data fail")
		return
	}

	if err = n.store.Set(key, value); err != nil {
		c.String(http.StatusInternalServerError, "set fail, err: %s", err)
		return
	}

	c.String(http.StatusOK, "ok")
}

func (n *node) get(c *gin.Context) {
	key := c.Query("key")

	value, err := n.store.Get(key)
	if err != nil {
		c.String(http.StatusNotFound, "get fail, err: %s", err)
	} else {
		c.String(http.StatusOK, value)
	}
}
