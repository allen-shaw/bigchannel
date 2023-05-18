package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		// Contact the server and print out its response.
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())

		cancel()
		time.Sleep(time.Second)
	}
}
