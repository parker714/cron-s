package tasks

type heap []*Task

func NewHeap() *heap {
	return &heap{}
}

func (h heap) Len() int           { return len(h) }
func (h heap) Less(i, j int) bool { return h[i].RunTime.Before(h[j].RunTime) }
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