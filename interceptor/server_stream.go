package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func StreamServerInterceptor(builder StreamServerInterceptorBuilder) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		meta := NewServerMeta(info.FullMethod, info, nil)
		intcptr, newCtx := builder.Build(ss.Context(), meta)
		intcptr.BeforeCreateStream(srv, info)
		err := handler(srv, &serverStream{ServerStream: ss, ctx: newCtx, intcptr: intcptr})
		intcptr.AfterCreateStream(srv, info, err)
		return err
	}
}

type serverStream struct {
	grpc.ServerStream
	ctx     context.Context
	intcptr streamServerInterceptor
}

func (ss *serverStream) SendMsg(m any) error {
	startTime := time.Now()
	ss.intcptr.BeforeSendMsg(m)
	err := ss.ServerStream.SendMsg(m)
	ss.intcptr.AfterSendMsg(m, err, time.Since(startTime))
	return err
}

func (ss *serverStream) RecvMsg(m any) error {
	startTime := time.Now()
	ss.intcptr.BeforeRecvMsg(m)
	err := ss.ServerStream.RecvMsg(m)
	ss.intcptr.AfterRecvMsg(m, err, time.Since(startTime))
	return err
}
