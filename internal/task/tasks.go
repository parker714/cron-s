package task

import (
	"container/heap"
)

// Tasks is app cached task data
type Tasks interface {
	heap.Interface
	Remove(name string)
	Top() interface{}
}

type tasks []*Task

// NewTasks returns Tasks instance
func NewTasks() Tasks {
	return &tasks{}
}

func (ts tasks) Len() int           { return len(ts) }
func (ts tasks) Less(i, j int) bool { return ts[i].PlanExecTime.Before(ts[j].PlanExecTime) }
func (ts tasks) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

func (ts *tasks) Push(x interface{}) {
	*ts = append(*ts, x.(*Task))
}

func (ts *tasks) Pop() interface{} {
	val := (*ts)[len(*ts)-1]
	*ts = (*ts)[:len(*ts)-1]
	return val
}

func (ts *tasks) Remove(name string) {
	for i, t := range *ts {
		if t.Name == name {
			heap.Remove(ts, i)
		}
	}
}

func (ts tasks) Top() interface{} {
	return ts[0]
}
