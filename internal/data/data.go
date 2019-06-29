package data

import (
	cheap "container/heap"
	"cron-s/internal/tasks"
	"sync"
)

type data struct {
	tasks *tasks.Heap
	mu    sync.RWMutex
}

var d *data

func init() {
	d = newData()
}

func newData() *data {
	return &data{
		tasks: tasks.NewHeap(),
	}
}

func Len() int { return d.Len() }
func (d *data) Len() int {
	d.mu.RLock()
	l := d.tasks.Len()
	d.mu.RUnlock()
	return l
}

func Add(t *tasks.Task) { d.Add(t) }
func (d *data) Add(t *tasks.Task) {
	d.mu.Lock()
	d.tasks.Push(t)
	d.mu.Unlock()
}

func Del(t *tasks.Task) { d.Add(t) }
func (d *data) Del(t *tasks.Task) {
	d.mu.Lock()
	for i, h := range *d.tasks {
		if t.Name == h.Name {
			cheap.Remove(d.tasks, i)
		}
	}
	d.mu.Unlock()
}

func Top() *tasks.Task { return d.Top() }
func (d *data) Top() *tasks.Task {
	d.mu.RLock()
	t := (*d.tasks)[0]
	d.mu.RUnlock()
	return t
}

func Fix(i int) { d.Fix(i) }
func (d *data) Fix(i int) {
	d.mu.Lock()
	cheap.Fix(d.tasks, i)
	d.mu.Unlock()
}

func All() *tasks.Heap { return d.All() }
func (d *data) All() *tasks.Heap {
	d.mu.RLock()
	t := d.tasks
	d.mu.RUnlock()

	return t
}

func Init(nh *tasks.Heap) { d.Init(nh) }
func (d *data) Init(nh *tasks.Heap) {
	d.mu.Lock()
	d.tasks = nh
	cheap.Init(d.tasks)
	d.mu.Unlock()
}
