package main

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	addr   string
	pb.BrokerClient
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	opts := prepareOpts()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		err = fmt.Errorf("dial server: %w", err)
		defer cancel(err)
		return nil, err
	}
	bc := pb.NewBrokerClient(conn)
	c := &Client{
		ctx:          ctx,
		cancel:       cancel,
		addr:         addr,
		BrokerClient: bc,
	}

	return c, nil
}

func (c *Client) Close() {
	c.cancel(errors.New("client close"))
}

func prepareOpts() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return opts
}
