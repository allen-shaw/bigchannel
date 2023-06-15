package interceptor

import (
	"context"
	"time"

	"github.com/allen-shaw/bigchannel/interceptor/meta"
	"google.golang.org/grpc"
)

func UnaryClientInterceptor(builder meta.UnaryClientInterceptorBuilder) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		meta := meta.NewClientMeta(method, nil, req)
		intcptr, newCtx := builder.Build(ctx, meta)
		startTime := time.Now()

		intcptr.BeforeRequest(req)
		err := invoker(newCtx, method, req, reply, cc, opts...)
		intcptr.AfterRequest(reply, err, time.Since(startTime))
		return err
	}
}
