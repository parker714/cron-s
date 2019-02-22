package server

import (
	"context"
	"cron-s/internal/task"
	"errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

type Etcd struct {
	Client            *clientv3.Client
	Kv                clientv3.KV
	Watcher           clientv3.Watcher
	Lease             clientv3.Lease
	WatchJobsRevision int64
}

func NewEtcd(endpoints []string) (e *Etcd, err error) {
	e = &Etcd{}
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
	}
	e.Client, err = clientv3.New(cfg)
	if err != nil {
		return
	}
	e.Kv = clientv3.NewKV(e.Client)
	e.Watcher = clientv3.NewWatcher(e.Client)
	e.Lease = clientv3.NewLease(e.Client)
	return
}

func (e *Etcd) Add(t *task.Task) error {
	ts, err := task.Marshal(t)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.TODO(), 3*time.Second)
	_, err = e.Kv.Put(ctx, JobsKey+t.Name, ts)

	return err
}

func (e *Etcd) Del(t *task.Task) error {
	ctx, _ := context.WithTimeout(context.TODO(), 3*time.Second)
	_, err := e.Kv.Delete(ctx, JobsKey+t.Name)

	return err
}

func (e *Etcd) Get(key string) (tasks map[string]*task.Task, err error) {
	tasks = make(map[string]*task.Task)

	ctx, _ := context.WithTimeout(context.TODO(), 3*time.Second)
	getResponse, err := e.Kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return
	}
	e.WatchJobsRevision = getResponse.Header.Revision + 1

	for _, v := range getResponse.Kvs {
		t, err := task.Unmarshal(v.Value)
		if err != nil {
			continue
		}
		tasks[t.Name] = t
	}

	return
}

func (e *Etcd) Watch(me chan *task.ModifyEvent) {
	watchChan := e.Watcher.Watch(context.TODO(), JobsKey, clientv3.WithRev(e.WatchJobsRevision), clientv3.WithPrefix())

	for w := range watchChan {
		for _, e := range w.Events {
			tmp := &task.ModifyEvent{
				Name: string(e.Kv.Key)[len(JobsKey):],
			}

			switch e.Type {
			case mvccpb.PUT:
				t, err := task.Unmarshal(e.Kv.Value)
				if err != nil {
					continue
				}
				tmp.Task = t
				tmp.Type = task.PUT
			case mvccpb.DELETE:
				tmp.Type = task.DEL
			}
			me <- tmp
		}
	}
}

func (e *Etcd) Lock(key string, do func()) (err error) {
	key += "lock"

	grant, err := e.Lease.Grant(context.TODO(), 3)
	if err != nil {
		return
	}
	ctx, cancelFunc := context.WithCancel(context.TODO())
	_, err = e.Lease.KeepAlive(ctx, grant.ID)
	if err != nil {
		return
	}
	defer cancelFunc()

	txnResp, err := e.Kv.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(grant.ID))).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return
	}
	if !txnResp.Succeeded {
		err = errors.New("task is locking")
		return
	}
	do()
	return
}

func (e *Etcd) Close() error {
	return e.Client.Close()
}
