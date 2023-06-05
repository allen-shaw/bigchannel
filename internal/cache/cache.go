package cache

import (
	"fmt"
	"sync"
)

type RingBuffer struct {
	// 数据
	data []int
	// ackCursor int

	ack  int
	len  int
	cap  int
	head int
	tail int

	mu   sync.Mutex
	cond *sync.Cond
}

func NewRingBuffer(size int) *RingBuffer {
	c := new(RingBuffer)
	c.data = make([]int, size)
	c.cap = size
	c.len = 0
	c.head = 0
	c.ack = 0
	c.tail = 0
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *RingBuffer) Push(in int) {
	c.mu.Lock()
	for c.len == c.cap {
		c.cond.Wait()
		fmt.Println("wait ok,", in)
	}
	fmt.Println("wait ok 1,", in)

	c.data[c.tail] = in
	c.tail++
	c.tail = c.tail % c.cap
	c.len++
	fmt.Println("wait ok 2", in)

	c.mu.Unlock()
	fmt.Println("wait ok 3", in)

	if c.len == 1 {
		c.cond.Signal()
	}
}

func (c *RingBuffer) Pop() (data int, pos int) {
	c.mu.Lock()

	for c.len == 0 {
		c.cond.Wait()
	}

	pos = c.head
	data = c.data[pos]
	c.head--
	c.head = c.head % c.cap
	c.mu.Unlock()

	return data, pos
}

func (c *RingBuffer) Ack(idx int) {
	c.mu.Lock()
	size := abs(c.ack - idx)
	c.ack = idx
	c.len -= size

	if c.len+size == c.cap {
		c.cond.Signal()
	}
	c.mu.Unlock()
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
