package task

import (
	cheap "container/heap"
	"github.com/pkg/errors"
)

// Heap is app cached task data
type Heap interface {
	cheap.Interface
	Top() (*Task, error)
	Remove(t *Task)
}

type heap []*Task

// NewHeap returns Heap instance
func NewHeap() Heap {
	return &heap{}
}

func (h heap) Len() int           { return len(h) }
func (h heap) Less(i, j int) bool { return h[i].PlanExecTime.Before(h[j].PlanExecTime) }
func (h heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *heap) Push(x interface{}) {
	*h = append(*h, x.(*Task))
}

func (h *heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *heap) Top() (*Task, error) {
	if len(*h) < 0 {
		return nil, errors.New("heap no data")
	}

	return (*h)[0], nil
}

func (h *heap) Remove(t *Task) {
	for i, hh := range *h {
		if t.Name == hh.Name {
			cheap.Remove(h, i)
		}
	}
}
