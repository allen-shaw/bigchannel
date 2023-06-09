package cache

import (
	"errors"
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
}

func NewRingBuffer(size int) *RingBuffer {
	c := new(RingBuffer)
	c.data = make([]int, size)
	c.cap = size
	c.len = 0
	c.head = -1
	c.ack = 0
	c.tail = -1
	return c
}

func (c *RingBuffer) Push(in int) error {
	if c.len == c.cap {
		return errors.New("buffer full")
	}

	c.tail++
	c.tail = c.tail % c.cap

	c.data[c.tail] = in
	c.len++
	return nil
}

func (c *RingBuffer) Pop() (data int, pos int, err error) {
	if c.len == 0 {
		return 0, 0, errors.New("buffer empty")
	}

	c.head++
	c.head = c.head % c.cap

	pos = c.head
	data = c.data[pos]

	return data, pos, nil
}

func (c *RingBuffer) Ack(idx int) {
	var size int
	if idx > c.ack {
		size = idx - c.ack + 1
	} else {
		size = c.ack + c.cap - idx + 1
	}

	c.ack = idx
	c.len -= size
}
