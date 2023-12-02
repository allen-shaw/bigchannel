package main

import (
	"sync"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
)

type Queue struct {
	cap  int
	mu   sync.RWMutex
	data []*sendingMsg
}

func newQueue(cap int) *Queue {
	q := &Queue{cap: cap}
	q.data = make([]*sendingMsg, 0, cap)
	return q
}

func (q *Queue) Push(m *sendingMsg) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.data = append(q.data, m)
}

func (q *Queue) Pop() *sendingMsg {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.data) == 0 {
		return nil
	}

	first := q.data[0]
	q.data = q.data[1:]
	return first
}

func (q *Queue) PopUntil(seqno uint64) []*sendingMsg {
	q.mu.Lock()
	defer q.mu.Unlock()

	idx := -1

	for id, msg := range q.data {
		if msg.m.SeqNo.SeqNo == seqno {
			idx = id + 1
			break
		}
	}

	if idx == -1 {
		return nil
	}

	msgs := q.data[:idx]
	q.data = q.data[idx:]

	return msgs
}

func (q *Queue) PopAll() []*sendingMsg {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := q.data
	q.data = make([]*sendingMsg, 0, q.cap)
	return out
}

func (q *Queue) Top() *sendingMsg {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.data) == 0 {
		return nil
	}

	return q.data[0]
}

func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.data)
}

func (q *Queue) Cap() int {
	return q.cap
}

type MsgQueue struct {
	cap  int
	mu   sync.RWMutex
	data []*pb.Message
}

func newMsgQueue(cap int) *MsgQueue {
	q := &MsgQueue{cap: cap}
	q.data = make([]*pb.Message, 0, cap)
	return q
}

func (q *MsgQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.data)
}

func (q *MsgQueue) Cap() int {
	return q.cap
}
func (q *MsgQueue) Push(msgs ...*pb.Message) {
	if len(msgs) == 0 {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()

	q.data = append(q.data, msgs...)
}

func (q *MsgQueue) Pop() *pb.Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.data) == 0 {
		return nil
	}

	first := q.data[0]
	q.data = q.data[1:]
	return first
}
