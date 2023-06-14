package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func StreamClientInterceptor(builder StreamClientInterceptorBuilder) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		meta := NewClientMeta(method, desc, nil)
		startTime := time.Now()
		intcptr, newCtx := builder.Build(ctx, meta)
		intcptr.BeforeCreateStream()
		cs, err := streamer(newCtx, desc, cc, method, opts...)
		if err != nil {
			intcptr.AfterCreateStream(err, time.Since(startTime))
			return nil, err
		}
		return &clientStream{ClientStream: cs, startTime: startTime, intcptr: intcptr}, nil
	}
}

type clientStream struct {
	grpc.ClientStream

	startTime time.Time
	intcptr   streamClientInterceptor
}

func (cs *clientStream) SendMsg(m any) error {
	startTime := time.Now()
	cs.intcptr.BeforeSendMsg(m)
	err := cs.ClientStream.SendMsg(m)
	cs.intcptr.AfterSendMsg(m, err, time.Since(startTime))
	return err
}

func (cs *clientStream) RecvMsg(m any) error {
	startTime := time.Now()
	cs.intcptr.BeforeRecvMsg(m)
	err := cs.ClientStream.RecvMsg(m)
	cs.intcptr.AfterRecvMsg(m, err, time.Since(startTime))

	cs.intcptr.AfterCreateStream(err, time.Since(cs.startTime))

	return err
}
