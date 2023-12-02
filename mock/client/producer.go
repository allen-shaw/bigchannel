package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
)

type Producer struct {
	id     []byte
	ctx    context.Context
	cancel context.CancelCauseFunc
	c      *Client
	stream *sendStream

	pendingQ *Queue
	sendingQ *Queue

	reqC  chan *pb.SendRequest
	respC chan *pb.SendResponse
}

func NewProducer(c *Client) (*Producer, error) {
	ctx, cancel := context.WithCancelCause(c.ctx)
	id := UUID()
	p := &Producer{
		id:       id,
		ctx:      ctx,
		cancel:   cancel,
		c:        c,
		pendingQ: newQueue(1024),
		sendingQ: newQueue(1024),
		reqC:     make(chan *pb.SendRequest, 1),
		respC:    make(chan *pb.SendResponse, 1),
	}
	err := p.init()
	if err != nil {
		err = fmt.Errorf("producer init: %w", err)
		p.cancel(err)
		return nil, err
	}

	return p, nil
}

func (p *Producer) init() error {
	ctx := p.ctx

	p.run()

	// 创建stream
	stream, err := p.c.Send(ctx)
	if err != nil {
		return fmt.Errorf("client send: %w", err)
	}

	ss := &sendStream{
		ctx:    p.ctx,
		stream: stream,
		reqC:   p.reqC,
		respC:  p.respC,
	}
	ss.run()
	p.stream = ss
	return nil
}

func (p *Producer) run() {
	go p.sendloop()
	go p.recvloop()
	go p.retryloop()
}

func (p *Producer) sendloop() {
	// 从sendingQ读取数据
	for {
		select {
		case <-p.ctx.Done():
			log.Printf("producer sendloop stop: %v \n", p.ctx.Err())
			return
		default:
		}

		// TODO: 这里应该实现成一个阻塞队列
		msg := p.sendingQ.Pop()
		if msg == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// 开始发送
		req := &pb.SendRequest{
			Messages: []*pb.Message{msg.m},
		}

		select {
		case <-p.ctx.Done():
			log.Printf("producer sendloop stop: %v \n", p.ctx.Err())
			return
		case p.reqC <- req:
			msg.status = StatusPendding
			p.pendingQ.Push(msg)
		}
	}
}

func (p *Producer) recvloop() {
	// 从sendingQ读取数据
	for {
		select {
		case <-p.ctx.Done():
			log.Printf("producer recvloop stop: %v \n", p.ctx.Err())
			return
		case resp, ok := <-p.respC:
			if !ok {
				log.Printf("producer recvloop stop: %v \n", fmt.Errorf("respC closed"))
				return
			}

			if !bytes.Equal(resp.SeqNo.ProducerId, p.id) {
				log.Printf("not my resp, resp_pid: %v != my.pid: %v", resp.SeqNo.ProducerId, p.id)
				continue
			}

			msgs := p.pendingQ.PopUntil(resp.SeqNo.SeqNo)
			for _, msg := range msgs {
				msg.markSucc()
			}
		}
	}
}

func (p *Producer) retryloop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("producer recvloop stop: %v \n", p.ctx.Err())
			return
		case <-ticker.C:
			for {
				// 检查penind.Q的top的发送时间
				pedingMsg := p.pendingQ.Top()
				if pedingMsg == nil {
					continue
				}

				d := time.Since(pedingMsg.lastSendTime)
				if d < 3*time.Second {
					continue
				}

				// 重试次数超过3次,队头超时，所有失败
				if pedingMsg.retryTimes >= 3 {
					msgs := p.pendingQ.PopAll()
					for _, msg := range msgs {
						msg.marFail()
					}
					continue
				}

				// resend
				req := &pb.SendRequest{
					Messages: []*pb.Message{pedingMsg.m},
				}

				select {
				case <-p.ctx.Done():
					log.Printf("producer sendloop stop: %v \n", p.ctx.Err())
					return
				case p.reqC <- req:
					_ = p.pendingQ.Pop()
					pedingMsg.lastSendTime = time.Now()
					pedingMsg.retryTimes++
					p.pendingQ.Push(pedingMsg)
				}

			}
		}
	}
}

// AsyncSend
func (p *Producer) Send(msg *pb.Message) error {
	p.sendingQ.Push(newSendingMsg(msg))
	return nil
}

type sendStream struct {
	ctx    context.Context
	stream pb.Broker_SendClient
	reqC   chan *pb.SendRequest
	respC  chan *pb.SendResponse
}

func (ss *sendStream) run() {
	go ss.sendloop()
	go ss.recvloop()
}

func (ss *sendStream) sendloop() {
	ctx := ss.stream.Context()
	for {
		select {
		case <-ss.ctx.Done():
			log.Printf("ss stop, with pctx done: %v \n", ss.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("ss stop, with stream ctx done: %v \n", ss.ctx.Err())
			return
		case req := <-ss.reqC: // FIXME: 这里要监听ok
			err := ss.stream.Send(req)
			if err != nil {
				// 这里合理是要处理重连的
				log.Printf("stream send: %v \n", err)
				return
			}
		}
	}
}

func (ss *sendStream) recvloop() {
	ctx := ss.stream.Context()
	for {
		select {
		case <-ss.ctx.Done():
			log.Printf("ss stop, with pctx ctx done: %v \n", ss.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("ss stop, with stream ctx done: %v \n", ss.ctx.Err())
			return
		default:
		}

		resp, err := ss.stream.Recv()
		if err != nil {
			log.Printf("stream recv: %v \n", err)
			return
		}

		select {
		case <-ss.ctx.Done():
			log.Printf("ss stop, with pctx ctx done: %v \n", ss.ctx.Err())
			return
		case <-ctx.Done():
			log.Printf("ss stop, with stream ctx done: %v \n", ss.ctx.Err())
			return
		case ss.respC <- resp:
		}
	}
}

type Status int8

const (
	StatusInit Status = iota
	StatusPendding
	StatusSucc
	StatusFail
)

// 应该叫ProduceMessage
type sendingMsg struct {
	m             *pb.Message
	firstSendTime time.Time
	lastSendTime  time.Time
	retryTimes    int
	status        Status
	waitC         chan struct{}
}

func newSendingMsg(m *pb.Message) *sendingMsg {
	now := time.Now()
	sm := &sendingMsg{
		m:             m,
		firstSendTime: now,
		lastSendTime:  now,
		retryTimes:    0,
		status:        StatusInit,
		waitC:         make(chan struct{}),
	}
	return sm
}

func (m *sendingMsg) Get() Status {
	if m.status == StatusSucc || m.status == StatusFail {
		return m.status
	}
	<-m.waitC
	return m.status
}

func (m *sendingMsg) marFail() {
	if m.status != StatusPendding && m.status != StatusInit {
		return
	}
	m.status = StatusFail
	select {
	case m.waitC <- struct{}{}:
	default:
	}
}

func (m *sendingMsg) markSucc() {
	if m.status != StatusPendding && m.status != StatusInit {
		return
	}
	m.status = StatusSucc
	select {
	case m.waitC <- struct{}{}:
	default:
	}
}
