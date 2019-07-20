package task

import (
	cheap "container/heap"
)

// Heap is app cached task data
type Heap interface {
	cheap.Interface
	Remove(name string)
	Top() interface{}
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
	val := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return val
}

func (h *heap) Remove(name string) {
	for i, hh := range *h {
		if t.Name == hh.Name {
			cheap.Remove(h, i)
		}
	}
}

func (h heap) Top() interface{} {
	return h[0]
}
