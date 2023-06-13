package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		meta := NewServerMeta(info.FullMethod, info, nil)

		err := handler(srv)
	}
}

type ServerStream struct {
	grpc.ServerStream

	ctx         context.Context
	interceptor Interceptor
}

func (ss *ServerStream) SendMsg(m any) error {
	startTime := time.Now()
	ss.interceptor.BeforeSendMsg(m)
	err := ss.ServerStream.SendMsg(m)
	ss.interceptor.AfterSendMsg(m, err, time.Since(startTime))
	return err
}


