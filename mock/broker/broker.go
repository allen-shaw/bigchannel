package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
)

var errMsgIDNotFound = errors.New("msgid not found")

type Broker struct {
	id    string
	r     *Receiver
	sub   *Subscription
	store *Storage

	*pb.UnimplementedBrokerServer
}

func NewBroker(addr string) *Broker {
	store := NewStorage()
	b := &Broker{
		id:    addr, // 临时使用addr作为id
		r:     NewReceiver(store),
		sub:   NewSubscription(store),
		store: store,
	}
	return b
}

// Ack implements pb.BrokerServer.
func (b *Broker) Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error) {
	log.Printf("[Broker.Ack]|%v", req.String())

	err := b.sub.Ack(req.MessageId)
	if err != nil {
		return nil, fmt.Errorf("sub ack: %w", err)
	}
	return &pb.AckResponse{}, nil
}

// Receive implements pb.BrokerServer.
func (b *Broker) Receive(stream pb.Broker_ReceiveServer) error {
	log.Printf("[Broker.Receive]")

	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("recv stream ctx done: %w", ctx.Err())
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			log.Printf("[Broker.Receive]|error %v", err)
			return fmt.Errorf("recv stream recv: %w", err)
		}
		log.Printf("[Broker.Receive]|%v", req.String())

		// 这个接口本质上拉消息，所以调用pullMessage接口
		// pullMessage 是一个block接口，如果没有数据，会阻塞指导有数据
		msgs, err := b.sub.pullMessages(req.MessageId, req.Size)
		if err != nil {
			// demo 简化实现，其实这里要分错误类型讨论, 不应该直接断掉连接
			return fmt.Errorf("pull msg: %w", err)
		}

		resp := &pb.ReceiveResponse{
			Messages: msgs,
		}
		err = stream.Send(resp)
		if err != nil {
			return fmt.Errorf("recv stream send resp: %w", err)
		}
	}
}

// Send implements pb.BrokerServer.
func (b *Broker) Send(stream pb.Broker_SendServer) error {
	log.Printf("[Broker.Send]")
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("send stream ctx done: %w", ctx.Err())
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			log.Printf("[Broker.Send]|error %v", err)
			return fmt.Errorf("send stream recv: %w", err)
		}
		// log.Printf("[Broker.Send]|%v", req.String())

		seqno, err := b.r.recvMessage(req.Messages...)
		if err != nil {
			return fmt.Errorf("receiver recv: %w", err)
		}

		resp := &pb.SendResponse{
			SeqNo: seqno,
		}
		err = stream.Send(resp)
		if err != nil {
			return fmt.Errorf("send stream send resp: %w", err)
		}
	}
}

type Receiver struct {
	id    uint64
	store *Storage
}

func NewReceiver(store *Storage) *Receiver {
	r := &Receiver{
		store: store,
	}
	return r
}

func (r *Receiver) recvMessage(msgs ...*pb.Message) (*pb.SequenceNo, error) {
	// 生成message id
	for _, msg := range msgs {
		msg.MessageId = make([]byte, 8)
		r.id++
		binary.LittleEndian.PutUint64(msg.MessageId, r.id)
	}

	err := r.store.AppendMessage(msgs...)
	if err != nil {
		return nil, fmt.Errorf("store append: %w", err)
	}

	lastMsg := msgs[len(msgs)-1]
	return lastMsg.SeqNo, nil
}

type Storage struct {
	mu   sync.Mutex
	data []*pb.Message // 临时写法，TODO:实现成阻塞队列
}

func NewStorage() *Storage {
	s := &Storage{
		data: make([]*pb.Message, 0, 1024),
	}
	return s
}

func (s *Storage) AppendMessage(msgs ...*pb.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, msgs...)
	return nil
}

func (s *Storage) Index(msgID []byte) int {
	if len(msgID) == 0 {
		return 0
	}
	// FIXME: 这里直接遍历，实际应该加map索引，或者二分查找
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, msg := range s.data {
		if bytes.Equal(msg.MessageId, msgID) {
			return id
		}
	}

	// FIXME: 实际上应该找最近的前一个id
	return -1
}

// 阻塞获取
func (s *Storage) ScanFrom(startIndex int, size int32) ([]*pb.Message, error) {
	for {
		msgs, err := s.scanFron(startIndex, size)
		if err != nil {
			return nil, err
		}
		if len(msgs) > 0 {
			// log.Printf("[Storage.ScanFrom]|return msg: %v", msgs)
			return msgs, nil
		}
		time.Sleep(3 * time.Second)
	}
}

func (s *Storage) scanFron(startIndex int, size int32) ([]*pb.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if startIndex > len(s.data) {
		return nil, nil
	}

	// log.Printf("[Storage.scanFrom]|data: %v, size: %v", s.data, len(s.data))
	endIndex := startIndex + int(size)
	if endIndex > len(s.data) {
		endIndex = len(s.data)
	}
	msgs := s.data[startIndex:endIndex]
	return msgs, nil
}

func (s *Storage) DeleteBefore(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = s.data[index:]
}

// consumer

type Dispatcher struct {
	s *Storage // 实际上这里应该委托给Scanner对象
}

func NewDispatcher(s *Storage) *Dispatcher {
	d := &Dispatcher{
		s: s,
	}
	return d
}

func (d *Dispatcher) pullMessages(start []byte, size int32) ([]*pb.Message, error) {
	// 找到start 的index, 如果start为nil，则startIndex = 0
	startIndex := d.s.Index(start)
	if startIndex == -1 {
		return nil, errMsgIDNotFound
	}

	// 然后获取size个数据，然后返回
	msgs, err := d.s.ScanFrom(startIndex+1, size)
	if err != nil {
		return nil, fmt.Errorf("dispatcher scan: %w", err)
	}

	return msgs, nil
}

type Subscription struct {
	mu     sync.RWMutex
	cursor []byte
	d      *Dispatcher
}

func NewSubscription(s *Storage) *Subscription {
	sub := &Subscription{
		d: NewDispatcher(s),
	}
	return sub
}

func (s *Subscription) Ack(msgID []byte) error {
	s.SetCursor(msgID)

	// 清理msg
	s.gc()
	return nil
}

func (s *Subscription) pullMessages(start []byte, size int32) ([]*pb.Message, error) {
	if start == nil {
		start = s.Cursor()
	}

	return s.d.pullMessages(start, size)
}

func (s *Subscription) Cursor() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	return bytes.Clone(s.cursor)
}

func (s *Subscription) SetCursor(cursor []byte) {
	c := bytes.Clone(cursor)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cursor = c
}

func (s *Subscription) gc() {
	cursor := s.Cursor()

	idx := s.d.s.Index(cursor)
	if idx == -1 {
		return
	}
	s.d.s.DeleteBefore(idx)
}
