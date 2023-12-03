package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroker(t *testing.T) {
	b := NewBroker("127.0.0.1:18088")
	lastSeqno, err := b.r.recvMessage(newMessages(5)...)
	assert.Nil(t, err)

	fmt.Println(lastSeqno.String())

	msgs, err := b.sub.pullMessages(make([]byte, 0), 3)
	assert.Nil(t, err)

	for _, msg := range msgs {
		fmt.Println(string(msg.Payload))
	}

	lastMsg := msgs[len(msgs)-1]
	msgs2, err := b.sub.pullMessages(lastMsg.MessageId, 2)
	assert.Nil(t, err)

	for _, msg := range msgs2 {
		fmt.Println(string(msg.Payload))
	}

}
