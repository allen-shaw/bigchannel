package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"errors"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
)

type Consumer struct {
	id     []byte
	ctx    context.Context
	cancel context.CancelCauseFunc
	c      *Client
	stream *recvStream

	//
	reqC  chan *pb.ReceiveRequest
	respC chan *pb.ReceiveResponse

	//
	cursor atomic.Value //  []byte // 记录拉取的最新的messageID
	//
	fmu   sync.RWMutex // fetchMutex
	msgC  chan *pb.Message
	recvQ *MsgQueue // 预拉取缓存

	// fetchRespC       chan struct{}
	lastPreFetchTime time.Time
}

func NewConsumer(c *Client) (*Consumer, error) {
	ctx, cancel := context.WithCancelCause(c.ctx)
	id := UUID()

	co := &Consumer{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		c:      c,
		recvQ:  newMsgQueue(16),
		msgC:   make(chan *pb.Message),
		reqC:   make(chan *pb.ReceiveRequest, 1),
		respC:  make(chan *pb.ReceiveResponse, 1),
		// fetchRespC: make(chan struct{}, 1),
	}

	err := co.init()
	if err != nil {
		err = fmt.Errorf("consumer init: %w", err)
		co.cancel(err)
		return nil, err
	}

	return co, nil
}

func (c *Consumer) init() error {
	ctx := c.ctx

	// 创建stream
	stream, err := c.c.Receive(ctx)
	if err != nil {
		return fmt.Errorf("client recv: %w", err)
	}

	rs := &recvStream{
		ctx:    c.ctx,
		stream: stream,
		reqC:   c.reqC,
		respC:  c.respC,
	}

	rs.run()
	c.stream = rs

	c.run()

	return nil
}

func (c *Consumer) run() {
	go c.sendloop()
	go c.recvloop()
	go c.prefetch()
}

func (c *Consumer) Receive(ctx context.Context) (*pb.Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("recv: %w", ctx.Err())
		case <-c.ctx.Done():
			return nil, fmt.Errorf("consumer closed: %w", ctx.Err())
		case msg, ok := <-c.msgC:
			if !ok {
				return nil, fmt.Errorf("consumer closed: msgC closed")
			}
			return msg, nil
		}
	}
}

func (c *Consumer) Ack(ctx context.Context, msg *pb.Message) error {
	req := &pb.AckRequest{
		MessageId: msg.MessageId,
	}
	// 可以直接往reqC 发送ack请求
	_, err := c.c.Ack(ctx, req)
	if err != nil {
		return fmt.Errorf("ack %w", err)
	}
	return nil
}

func (c *Consumer) sendloop() {
	for {
		select {
		case <-c.ctx.Done():
			log.Printf("consumer sendloop stop: %v \n", c.ctx.Err())
			return
		default:
		}

		// 从c.recvQ读取数据
		for {
			msg := c.recvQ.Pop()
			if msg == nil {
				time.Sleep(time.Second)
				break
			}

			log.Printf("recvQ pop %v", msg)
			select {
			case <-c.ctx.Done():
				log.Printf("consumer sendloop stop: %v \n", c.ctx.Err())
				return

			case c.msgC <- msg:
			}
		}
	}
}

func (c *Consumer) recvloop() {
	for {
		select {
		case <-c.ctx.Done():
			log.Printf("consumer recvloop stop: %v \n", c.ctx.Err())
			return
		case resp, ok := <-c.respC:
			if !ok {
				// log.Printf("consumer recvloop stop: %v \n", fmt.Errorf("respC closed"))
				return
			}
			if len(resp.Messages) == 0 {
				continue
			}
			log.Printf("[Consumer.recvloop]|recv:%v", resp.Messages)
			lstID := c.recvQ.Push(resp.Messages...)
			log.Printf("[RecvQ Return]|%v", binary.LittleEndian.Uint64(lstID))
			c.SetCursor(lstID)
			// <-c.fetchRespC
		}
	}
}

// func (c *Consumer) prefetchAsync() {
// 	if time.Since(c.lastPreFetchTime) < 3*time.Second {
// 		return
// 	}
// 	go c.prefetch()
// 	c.lastPreFetchTime = time.Now()
// }

func (c *Consumer) prefetch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		// case c.fetchRespC <- struct{}{}:
		// 	c.fetch()
		case <-ticker.C:
			c.fetch()
		}
	}
}

func (c *Consumer) fetch() {
	c.fmu.Lock()
	defer c.fmu.Unlock()

	// recvQ 已经很大，无需fetch
	cap := c.recvQ.Cap()
	size := c.recvQ.Size()

	log.Printf("[Consumer.fetch]|cap: %v, size: %v", cap, size)
	if size*2 >= cap {
		return
	}

	req := &pb.ReceiveRequest{
		MessageId: c.Cursor(),
		Size:      int32(cap - size),
	}
	select {
	case c.reqC <- req:
		log.Printf("[Consumer.fetch]send request: %v", req.String())
	default:
	}
}

func (c *Consumer) Cursor() []byte {
	cur := c.cursor.Load()
	if cur == nil {
		return nil
	}
	return cur.([]byte)
}

func (c *Consumer) SetCursor(cursor []byte) {
	c.cursor.Store(cursor)
}

func (c *Consumer) Close() {
	c.stream.Close()
	c.cancel(errors.New("consuemr close"))
}

type recvStream struct {
	ctx    context.Context
	stream pb.Broker_ReceiveClient
	reqC   chan *pb.ReceiveRequest
	respC  chan *pb.ReceiveResponse
}

func (rs *recvStream) run() {
	go rs.sendloop()
	go rs.recvloop()
}

func (rs *recvStream) sendloop() {
	ctx := rs.stream.Context()
	for {
		select {
		case <-rs.ctx.Done():
			log.Printf("rs stop, with cctx done: %v \n", rs.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("rs stop, with stream ctx done: %v \n", ctx.Err())
			return
		case req, ok := <-rs.reqC:
			if !ok {
				log.Printf("rs stop, reqC closed")
				return
			}
			err := rs.stream.Send(req)
			if err != nil {
				// 这里合理是要处理重连的
				log.Printf("stream send: %v \n", err)
				return
			}
		}
	}

}

func (rs *recvStream) recvloop() {
	ctx := rs.stream.Context()
	for {
		select {
		case <-rs.ctx.Done():
			log.Printf("rs stop, with cctx ctx done: %v \n", rs.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("rs stop, with stream ctx done: %v \n", ctx.Err())
			return
		default:
		}

		resp, err := rs.stream.Recv()
		if err != nil {
			log.Printf("stream recv: %v \n", err)
			return
		}

		select {
		case <-rs.ctx.Done():
			log.Printf("rs stop, with cctx ctx done: %v \n", rs.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("rs stop, with stream ctx done: %v \n", ctx.Err())
			return
		case rs.respC <- resp:
		}
	}
}

func (rs *recvStream) Close() {
	rs.stream.CloseSend()
}
