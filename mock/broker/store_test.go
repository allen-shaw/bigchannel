package main

import (
	"encoding/binary"
	"fmt"
	"testing"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	store := NewStorage()

	msgs := newMessages(5)
	store.AppendMessage(msgs...)

	nilIndex := store.Index(make([]byte, 0))
	assert.Equal(t, 0, nilIndex)

	oneIndex := store.Index(msgs[1].MessageId)
	assert.Equal(t, 1, oneIndex)

	ms, _ := store.ScanFrom(4, 10)
	fmt.Println()
	for _, m := range ms {
		fmt.Println(m.String())
	}
}

func newMessages(size int) []*pb.Message {
	msgs := make([]*pb.Message, 0, size)
	for i := 0; i < size; i++ {
		mid := make([]byte, 8)
		binary.LittleEndian.PutUint64(mid, uint64(i))
		msg := &pb.Message{
			MessageId: mid,
			SeqNo: &pb.SequenceNo{
				ProducerId: []byte("pid0"),
				SeqNo:      uint64(i),
			},
			Payload: []byte(fmt.Sprintf("hello-%v", i)),
		}
		msgs = append(msgs, msg)
	}
	return msgs
}
