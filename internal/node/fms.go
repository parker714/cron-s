package node

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
	"log"
)

type logEntryData struct {
	Key   string
	Value string
}

// node运行时, 处理leader节点下发的数据（leader节点在应用层调用raft.Apply方法）
func (n *node) Apply(l *raft.Log) interface{}  {
	log.Println("node apply data, ", l.Data)

	led := &logEntryData{}
	if err := json.Unmarshal(l.Data, led); err != nil {
		log.Printf("node apply log entry data Unmarshal fail, err: %s", err)
		return nil
	}

	return n.store.Set(led.Key, led.Value)
}

// node运行时, 定时调用, 创建快照
// rc.SnapshotInterval   时间间隔
// rc.SnapshotThreshold  日志数目
func (n *node) Snapshot() (raft.FSMSnapshot, error) {
	log.Println("node start create snapshot")

	// 继续调用fmsSnapshot的（Persist、Release）方法
	return n, nil
}

// node运行时, 快照创建
// Persist应该将所有必要的状态转储到WriteCloser'sink'， 并在完成时调用sink.Close（）或在出错时调用sink.Cancel（）
func (n *node) Persist(sink raft.SnapshotSink) error {
	log.Println("node fmsSnapshot Persist")

	snapshotBytes, err := n.store.Encode()
	if err != nil {
		_ = sink.Cancel()
		return err
	}
	if _, err := sink.Write(snapshotBytes); err != nil {
		_ = sink.Cancel()
		return err
	}
	if err := sink.Close(); err != nil {
		_ = sink.Cancel()
		return err
	}

	return nil
}

// node运行时, 快照创建完成
func (n *node) Release() {
	log.Println("node fmsSnapshot Release")
}

// node启动时, 从本地日志恢复数据
func (n *node) Restore(serialized io.ReadCloser) error {
	log.Println("node restore local data")
	//var newData map[string]string
	//if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
	//	return err
	//}
	//
	//f.c.data = newData
	return nil
}
