package notifier

import (
	"sync"
)

type Notifier[T any] struct {
	mu sync.RWMutex
	cs map[string]chan T
}

func New[T any]() *Notifier[T] {
	n := &Notifier[T]{
		cs: make(map[string]chan T, 0),
	}
	return n
}

func (n *Notifier[T]) Notify(event T) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, c := range n.cs {
		n.notify(c, event)
	}
}

func (n *Notifier[T]) Listen(id string) <-chan T {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch, ok := n.cs[id]
	if ok {
		return ch
	}

	ch = make(chan T)
	n.cs[id] = ch
	return ch
}

func (n *Notifier[T]) notify(ch chan T, e T) {
	select {
	case ch <- e:
		return
	default:
	}
}
