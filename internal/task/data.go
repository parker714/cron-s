package task

import (
	cheap "container/heap"
	"sync"
)

type Data struct {
	tasks *heap
	mu    sync.RWMutex
}

func NewData() *Data {
	return &Data{
		tasks: NewHeap(),
	}
}

func (d *Data) Len() int {
	d.mu.RLock()
	l := d.tasks.Len()
	d.mu.RUnlock()
	return l
}

func (d *Data) Add(t *Task) {
	d.mu.Lock()
	d.tasks.Push(t)
	d.mu.Unlock()
}

func (d *Data) Del(t *Task) {
	d.mu.Lock()
	for i, h := range *d.tasks {
		if t.Name == h.Name {
			cheap.Remove(d.tasks, i)
		}
	}
	d.mu.Unlock()
}

func (d *Data) Top() *Task {
	d.mu.RLock()
	t := (*d.tasks)[0]
	d.mu.RUnlock()
	return t
}

func (d *Data) Fix(i int) {
	d.mu.Lock()
	cheap.Fix(d.tasks, i)
	d.mu.Unlock()
}

func (d *Data) All() *heap {
	d.mu.RLock()
	t := d.tasks
	d.mu.RUnlock()

	return t
}

func (d *Data) Init(nh *heap) {
	d.mu.Lock()
	d.tasks = nh
	cheap.Init(d.tasks)
	d.mu.Unlock()
}
