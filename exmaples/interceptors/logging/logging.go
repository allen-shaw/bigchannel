package logging

import (
	"context"
	"log"
	"time"

	"github.com/allen-shaw/bigchannel/interceptor"
	"github.com/allen-shaw/bigchannel/interceptor/meta"
	"google.golang.org/grpc"
)

type LoggingStreamServerIntercetorBuilder struct{}

type loggingStreamServerInterceptor struct {
}

func (l *loggingStreamServerInterceptor) BeforeSendMsg(m any) {
	log.Printf("before send msg, req: %v", m)
}
func (l *loggingStreamServerInterceptor) AfterSendMsg(m any, err error, d time.Duration) {
	log.Printf("after send msg, req: %v, err %v, timecosr %v", m, err, d)
}
func (l *loggingStreamServerInterceptor) BeforeRecvMsg(m any) {
	log.Printf("before recv msg, req: %v", m)
}
func (l *loggingStreamServerInterceptor) AfterRecvMsg(m any, err error, d time.Duration) {
	log.Printf("after recv msg, req: %v, err %v, timecosr %v", m, err, d)
}
func (l *loggingStreamServerInterceptor) BeforeCreateStream(srv any, info *grpc.StreamServerInfo) {
	log.Printf("before create stream, srv: %v, method: %v", srv, info.FullMethod)
}
func (l *loggingStreamServerInterceptor) AfterCreateStream(srv any, info *grpc.StreamServerInfo, err error) {
	log.Printf("after create stream, srv: %v, method: %v, err %v", srv, info.FullMethod, err)
}

func (b *LoggingStreamServerIntercetorBuilder) Build(ctx context.Context, meta meta.Meta) (meta.StreamServerInterceptor, context.Context) {
	lssi := new(loggingStreamServerInterceptor)
	return lssi, ctx
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return interceptor.StreamServerInterceptor(&LoggingStreamServerIntercetorBuilder{})
}
