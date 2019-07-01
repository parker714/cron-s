package util

import (
	"sync"
)

// WaitGroupWrapper struct
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// Wrap close app goroutine when app exit
func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
