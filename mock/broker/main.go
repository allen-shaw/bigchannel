package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/allen-shaw/bigchannel/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	StartBroker(18088)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-sig
}

type RPCServer struct {
	srv  *grpc.Server
	addr string
	b    *Broker
}

func NewRPCServer(port int32) *RPCServer {
	addr := fmt.Sprintf(":%v", port)

	opts := prepareOpts()
	s := &RPCServer{
		srv:  grpc.NewServer(opts...),
		addr: addr,
	}
	b := NewBroker(addr)
	pb.RegisterBrokerServer(s.srv, b)
	reflection.Register(s.srv)

	s.b = b
	return s
}

func prepareOpts() []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)
	return opts
}

func (s *RPCServer) Start() {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(fmt.Errorf("listen tcp: %w", err))
	}
	go func() {
		err := s.srv.Serve(lis)
		if err != nil {
			panic(fmt.Errorf("grpc serve: %w", err))
		}
	}()
}

func StartBroker(port int32) *RPCServer {
	s := NewRPCServer(port)
	s.Start()
	fmt.Printf("broker start at: %v \n", port)
	return s
}
