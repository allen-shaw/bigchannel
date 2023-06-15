package meta

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type StreamInterceptor interface {
	BeforeSendMsg(m any)
	AfterSendMsg(m any, err error, d time.Duration)
	BeforeRecvMsg(m any)
	AfterRecvMsg(m any, err error, d time.Duration)
}

type StreamServerInterceptor interface {
	StreamInterceptor
	BeforeCreateStream(srv any, info *grpc.StreamServerInfo)
	AfterCreateStream(srv any, info *grpc.StreamServerInfo, err error) // TODO: 应该是stream close
}

type StreamClientInterceptor interface {
	StreamInterceptor
	BeforeCreateStream()
	AfterCreateStream(err error, d time.Duration)
}

type UnaryServerInterceptor interface {
	BeforeRecv(req any)
	AfterRecv(resp any, err error, d time.Duration)
}

type UnaryClientInterceptor interface {
	BeforeRequest(req any)
	AfterRequest(resp any, err error, d time.Duration)
}

type StreamServerInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (StreamServerInterceptor, context.Context)
}

type StreamClientInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (StreamClientInterceptor, context.Context)
}

type UnaryServerInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (UnaryServerInterceptor, context.Context)
}

type UnaryClientInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (UnaryClientInterceptor, context.Context)
}
