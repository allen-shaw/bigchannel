package interceptor

import (
	"strings"

	"google.golang.org/grpc"
)

type GRPCType string

const (
	GRPCUnary        GRPCType = "unary"
	GRPCClientStream GRPCType = "client_stream"
	GRPCServerStream GRPCType = "server_stream"
	GRPCBidiStream   GRPCType = "bidi_stream"
)

type Meta struct {
	Req      any
	Type     GRPCType
	Service  string
	Method   string
	IsClient bool
}

func (m Meta) FullMethod() string {
	return "/" + m.Service + "/" + m.Method
}

func NewClientMeta(fullMethod string, streamDesc *grpc.StreamDesc, req any) Meta {
	c := Meta{}
	c.Req = req
	c.Type = clientStreamType(streamDesc)
	c.Service, c.Method = splitFullMethod(fullMethod)
	c.IsClient = true
	return c
}

func NewServerMeta(fullMethod string, streamInfo *grpc.StreamServerInfo, req any) Meta {
	c := Meta{}
	c.Req = req
	c.Type = serverStreamType(streamInfo)
	c.Service, c.Method = splitFullMethod(fullMethod)
	c.IsClient = false
	return c
}

func clientStreamType(desc *grpc.StreamDesc) GRPCType {
	if desc.ClientStreams && !desc.ServerStreams {
		return GRPCClientStream
	} else if !desc.ClientStreams && desc.ServerStreams {
		return GRPCServerStream
	}
	return GRPCBidiStream
}

func serverStreamType(info *grpc.StreamServerInfo) GRPCType {
	if info.IsClientStream && !info.IsServerStream {
		return GRPCClientStream
	} else if !info.IsClientStream && info.IsServerStream {
		return GRPCServerStream
	}
	return GRPCBidiStream
}

func splitFullMethod(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/") // remove leading slash
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", "unknown"
}
