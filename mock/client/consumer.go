package main

import (
	"context"
	"fmt"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
)

type Consumer struct {
	id     []byte
	ctx    context.Context
	cancel context.CancelCauseFunc
	c      *Client
	stream *recvStream

	//
	recvQ *Queue // 预拉取缓存
}

func NewConsumer(c *Client) (*Consumer, error) {
	ctx, cancel := context.WithCancelCause(c.ctx)
	id := UUID()

	co := &Consumer{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		c:      c,
		recvQ:  newQueue(64),
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
	}

	rs.run()
	c.stream = rs

	c.run()

	return nil
}

func (c *Consumer) run() {
	go c.sendloop()
	go c.recvloop()
}

func (c *Consumer) Receive(ctx context.Context) (*pb.Message, error) {
	panic("no implement")
}

func (c *Consumer) Ack(ctx context.Context, msg *pb.Message) error {
	panic("no implement")
}

func (c *Consumer) sendloop() {
	panic("no implement")
}

func (c *Consumer) recvloop() {
	panic("no implement")
}

type recvStream struct {
	ctx    context.Context
	stream pb.Broker_ReceiveClient
}

func (rs *recvStream) run() {
	go rs.sendloop()
	go rs.recvloop()
}

func (rs *recvStream) sendloop() {}

func (rs *recvStream) recvloop() {}
