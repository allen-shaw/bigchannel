package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/allen-shaw/bigchannel/client/picker"
	"github.com/allen-shaw/bigchannel/client/resolver"
	pb "github.com/allen-shaw/bigchannel/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	name = flag.String("name", defaultName, "Name to greet")

	myScheme      = resolver.Scheme
	myServiceName = resolver.ServiceName
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:///%s", myScheme, myServiceName),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, picker.Scheme)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// for {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// 	// Contact the server and print out its response.
	// 	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	// 	if err != nil {
	// 		log.Fatalf("could not greet: %v", err)
	// 	}
	// 	log.Printf("Greeting: %s", r.GetMessage())

	// 	cancel()
	// 	time.Sleep(time.Second)
	// }
	sayHello3(c)
}

func sayHello3(c pb.GreeterClient) error {

	stream, err := c.SayHello3(context.Background())
	if err != nil {
		log.Printf("failed to call: %v", err)
		return err
	}

	var (
		wg         sync.WaitGroup
		rerr, serr error
	)

	reqC := make(chan string, 100)
	respC := make(chan string, 100)

	wg.Add(3)
	go func() {
		defer wg.Done()
		rerr = recvloop(stream, respC)
		fmt.Println("************** recvloop out")
	}()

	go func() {
		defer wg.Done()
		serr = sendloop(stream, reqC)
		fmt.Println("************** sendloop out")
	}()

	go func() {
		defer wg.Done()
		start(reqC, respC)
		fmt.Println("************** start out")
	}()
	wg.Wait()

	close(respC)
	if rerr != nil || serr != nil {
		return fmt.Errorf("rerr: %v, serr %v", rerr, serr)
	}

	return nil
}

func start(reqC, respC chan string) {
	go func() {
		for {
			resp, ok := <-respC
			if !ok {
				return
			}
			log.Println(resp)
		}
	}()

	for i := 0; i < 5; i++ {
		reqC <- strconv.Itoa(i)
		time.Sleep(time.Second)
	}
	close(reqC)
}

func recvloop(stream pb.Greeter_SayHello3Client, dataC chan<- string) error {
	ctx := stream.Context()
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
		case dataC <- reply.Message:
		}
	}
}

func sendloop(stream pb.Greeter_SayHello3Client, dataC <-chan string) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			stream.CloseSend() // TODO: test
			return ctx.Err()
		case req, ok := <-dataC:
			if !ok {
				stream.CloseSend()
				return nil
			}
			err := stream.Send(&pb.HelloRequest{Name: *name + req})
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
