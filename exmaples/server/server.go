package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/allen-shaw/bigchannel/exmaples/interceptors/logging"
	pb "github.com/allen-shaw/bigchannel/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName() + fmt.Sprintf(":%d", *port)}, nil
}

func (s *server) SayHello3(stream pb.Greeter_SayHello3Server) error {
	var (
		wg         sync.WaitGroup
		rerr, serr error
	)

	dataC := make(chan string, 100)

	wg.Add(2)
	go func() {
		defer wg.Done()
		rerr = recvloop(stream, dataC)
		fmt.Println("************** recvloop out")
	}()

	go func() {
		defer wg.Done()
		serr = sendloop(stream, dataC)
		fmt.Println("************** sendloop out")
	}()

	wg.Wait()

	if rerr != nil || serr != nil {
		return fmt.Errorf("rerr: %v, serr %v", rerr, serr)
	}

	return nil
}

func recvloop(stream pb.Greeter_SayHello3Server, dataC chan<- string) error {
	ctx := stream.Context()
	defer close(dataC)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		reply, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Printf("failed to recv: %v", err)
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataC <- reply.Name:
		}
	}

}

func sendloop(stream pb.Greeter_SayHello3Server, dataC <-chan string) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req, ok := <-dataC:
			if !ok {
				return nil
			}
			err := stream.Send(&pb.HelloReply{Message: "Hello " + req})
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Printf("failed to send: %v", err)
				return err
			}
		}
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	intcptrs := prepareInterceptors()
	s := grpc.NewServer(intcptrs)
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func prepareInterceptors() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(
		logging.StreamServerInterceptor(),
		logging.StreamServerInterceptor(),
	)
}
