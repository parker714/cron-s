package tasks

import (
	cheap "container/heap"
	"sync"
)

var (
	mu sync.RWMutex
	ts *heap
)

func init() {
	ts = &heap{}
}

func Len() int {
	mu.RLock()
	l := ts.Len()
	mu.RUnlock()
	return l
}

func Add(t *Task) {
	mu.Lock()
	ts.Push(t)
	mu.Unlock()
}

func Del(t *Task) {
	mu.Lock()
	for i, h := range *ts {
		if t.Name == h.Name {
			cheap.Remove(ts, i)
		}
	}
	mu.Unlock()
}

func Top() *Task {
	mu.RLock()
	t := (*ts)[0]
	mu.RUnlock()
	return t
}

func Fix(i int) {
	mu.Lock()
	cheap.Fix(ts, i)
	mu.Unlock()
}

func All() *heap {
	mu.RLock()
	t := ts
	mu.RUnlock()

	return t
}

func Init(nh *heap) {
	mu.Lock()
	ts = nh
	cheap.Init(ts)
	mu.Unlock()
}
