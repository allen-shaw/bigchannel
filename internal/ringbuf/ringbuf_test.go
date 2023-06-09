package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sk-CwWepovYHst5jQIt5Wl3T3BlbkFJpVw2ekscHIJcZQCV

func TestRingBuffer(t *testing.T) {
	buf := NewRingBuffer(10)
	var err error

	for i := 0; i < 10; i++ {
		err = buf.Push(i)
		assert.Nil(t, err)
	}
	fmt.Println("1-", "head:", buf.head, "tail:", buf.tail, "ack:", buf.ack, "len:", buf.len, "data:", buf.data)

	err = buf.Push(11)
	assert.NotNil(t, err)

	var (
		v, pos int
	)
	for i := 0; i < 2; i++ {
		v, pos, err = buf.Pop()
		assert.Nil(t, err)
		fmt.Println("2-", i, "head:", buf.head, "tail:", buf.tail, "ack:", buf.ack, "len:", buf.len, "data:", buf.data)
	}

	fmt.Println("pos:", pos, "val:", v)
	buf.Ack(pos)

	fmt.Println("3-", "head:", buf.head, "tail:", buf.tail, "ack:", buf.ack, "len:", buf.len, "data:", buf.data)
	err = buf.Push(11)
	assert.Nil(t, err)
	err = buf.Push(12)
	assert.Nil(t, err)
	err = buf.Push(13)
	assert.NotNil(t, err)

	v, pos, err = buf.Pop()
	assert.Nil(t, err)
	fmt.Println("pos:", pos, "val:", v)

	buf.Ack(pos)

	err = buf.Push(13)
	assert.Nil(t, err)

	fmt.Println("4-", "head:", buf.head, "tail:", buf.tail, "ack:", buf.ack, "len:", buf.len, "data:", buf.data)
}
