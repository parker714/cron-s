package tasks

type Heap []*Task

func NewHeap() *Heap {
	return &Heap{}
}

func (h Heap) Len() int           { return len(h) }
func (h Heap) Less(i, j int) bool { return h[i].RunTime.Before(h[j].RunTime) }
func (h Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *Heap) Push(x interface{}) {
	*h = append(*h, x.(*Task))
}

func (h *Heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
