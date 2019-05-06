package crond

import (
	cheap "container/heap"
	"cron-s/tasks"
	"sync"
)

type taskManage struct {
	data *tasks.Heap
	mu   sync.RWMutex
}

func newManage() *taskManage {
	return &taskManage{
		data: tasks.NewHeap(),
	}
}

func (tm *taskManage) Len() int {
	tm.mu.RLock()
	l := tm.data.Len()
	tm.mu.RUnlock()
	return l
}

func (tm *taskManage) Add(t *tasks.Task) {
	tm.mu.Lock()
	tm.data.Push(t)
	tm.mu.Unlock()
}

func (tm *taskManage) Del(t *tasks.Task) {
	tm.mu.Lock()
	for i, h := range *tm.data {
		if t.Name == h.Name {
			cheap.Remove(tm.data, i)
		}
	}
	tm.mu.Unlock()
}

func (tm *taskManage) Top() *tasks.Task {
	tm.mu.RLock()
	t := (*tm.data)[0]
	tm.mu.RUnlock()
	return t
}

func (tm *taskManage) Fix(i int) {
	tm.mu.Lock()
	cheap.Fix(tm.data, i)
	tm.mu.Unlock()
}

func (tm *taskManage) All() *tasks.Heap {
	tm.mu.RLock()
	t := tm.data
	tm.mu.RUnlock()

	return t
}

func (tm *taskManage) Init(nh *tasks.Heap) {
	tm.mu.Lock()
	tm.data = nh
	cheap.Init(tm.data)
	tm.mu.Unlock()
}
