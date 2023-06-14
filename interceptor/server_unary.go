package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(builder UnaryServerInterceptorBuilder) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		meta := NewServerMeta(info.FullMethod, nil, req)
		intcptr, newCtx := builder.Build(ctx, meta)
		startTime := time.Now()

		intcptr.BeforeRecv(req)
		resp, err = handler(newCtx, req)
		intcptr.AfterRecv(resp, err, time.Since(startTime))
		return resp, err
	}
}
