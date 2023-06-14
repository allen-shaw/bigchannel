package interceptor

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

type streamServerInterceptor interface {
	StreamInterceptor
	BeforeCreateStream(srv any, info *grpc.StreamServerInfo)
	AfterCreateStream(srv any, info *grpc.StreamServerInfo, err error)
}

type streamClientInterceptor interface {
	StreamInterceptor
	BeforeCreateStream()
	AfterCreateStream(err error, d time.Duration)
}

type unaryServerInterceptor interface {
	BeforeRecv(req any)
	AfterRecv(resp any, err error, d time.Duration)
}

type unaryClientInterceptor interface {
	BeforeRequest(req any)
	AfterRequest(resp any, err error, d time.Duration)
}

type StreamServerInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (streamServerInterceptor, context.Context)
}

type StreamClientInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (streamClientInterceptor, context.Context)
}

type UnaryServerInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (unaryServerInterceptor, context.Context)
}

type UnaryClientInterceptorBuilder interface {
	Build(ctx context.Context, meta Meta) (unaryClientInterceptor, context.Context)
}
