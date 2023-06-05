package cache

import (
	"fmt"
	"testing"
	"time"
)

// sk-CwWepovYHst5jQIt5Wl3T3BlbkFJpVw2ekscHIJcZQCV

func TestRingBuffer(t *testing.T) {
	buf := NewRingBuffer(10)

	for i := 0; i < 10; i++ {
		buf.Push(i)
	}

	fmt.Println(buf.data)
	go func() {
		fmt.Println("before block")
		buf.Push(11)
		fmt.Println("after block", buf.data)
	}()

	time.Sleep(5 * time.Second)
	v, pos := buf.Pop()
	fmt.Println("pos:", pos, "val:", v)
	buf.Ack(pos)
	fmt.Println("123", buf.data)
	time.Sleep(5 * time.Second)

}
