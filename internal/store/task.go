package store

import (
	cheap "container/heap"
	"cron-s/pkg/tasks"
	"sync"
)

type task struct {
	data *tasks.Heap
	mu   sync.RWMutex
}

var tm *task

func init() {
	tm = NewTask()
}

func NewTask() *task {
	return &task{
		data: tasks.NewHeap(),
	}
}

func Len() int { return tm.Len() }
func (tm *task) Len() int {
	tm.mu.RLock()
	l := tm.data.Len()
	tm.mu.RUnlock()
	return l
}

func Add(t *tasks.Task) { tm.Add(t) }
func (tm *task) Add(t *tasks.Task) {
	tm.mu.Lock()
	tm.data.Push(t)
	tm.mu.Unlock()
}

func Del(t *tasks.Task) { tm.Add(t) }
func (tm *task) Del(t *tasks.Task) {
	tm.mu.Lock()
	for i, h := range *tm.data {
		if t.Name == h.Name {
			cheap.Remove(tm.data, i)
		}
	}
	tm.mu.Unlock()
}

func Top() *tasks.Task { return tm.Top() }
func (tm *task) Top() *tasks.Task {
	tm.mu.RLock()
	t := (*tm.data)[0]
	tm.mu.RUnlock()
	return t
}

func Fix(i int) { tm.Fix(i) }
func (tm *task) Fix(i int) {
	tm.mu.Lock()
	cheap.Fix(tm.data, i)
	tm.mu.Unlock()
}

func All() *tasks.Heap { return tm.All() }
func (tm *task) All() *tasks.Heap {
	tm.mu.RLock()
	t := tm.data
	tm.mu.RUnlock()

	return t
}

func Init(nh *tasks.Heap) { tm.Init(nh) }
func (tm *task) Init(nh *tasks.Heap) {
	tm.mu.Lock()
	tm.data = nh
	cheap.Init(tm.data)
	tm.mu.Unlock()
}
