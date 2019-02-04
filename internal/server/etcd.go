package server

import (
	"context"
	"cron-s/internal/job"
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

func (e *Etcd) Get() (jobs map[string]*job.Job, err error) {
	jobs = make(map[string]*job.Job)

	getResponse, err := e.Kv.Get(context.TODO(), JobsKey, clientv3.WithPrefix())
	if err != nil {
		return
	}
	e.WatchJobsRevision = getResponse.Header.Revision + 1

	for _, v := range getResponse.Kvs {
		j, err := job.Unmarshal(v.Value)
		if err != nil {
			continue
		}
		jobs[string(v.Key)] = j
	}

	return
}

func (e *Etcd) Watch(jc chan *job.ChangeEvent) {
	watchChan := e.Watcher.Watch(context.TODO(), JobsKey, clientv3.WithRev(e.WatchJobsRevision), clientv3.WithPrefix())

	for w := range watchChan {
		for _, e := range w.Events {
			tmp := &job.ChangeEvent{
				Key: string(e.Kv.Key),
			}

			switch e.Type {
			case mvccpb.PUT:
				j, err := job.Unmarshal(e.Kv.Value)
				if err != nil {
					continue
				}
				tmp.Job = j
				tmp.Type = job.PUT
			case mvccpb.DELETE:
				tmp.Type = job.DEL
			}

			jc <- tmp
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
		err = errors.New("job is lock")
		return
	}
	do()
	return
}
