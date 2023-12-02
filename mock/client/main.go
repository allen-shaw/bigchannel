package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	addr := "127.0.0.1:8080"
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}

	startProducer(client)
	startConsumer(client)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-sig
}

func startProducer(c *Client) {
	producer, err := NewProducer(c)
	if err != nil {
		panic(err)
	}

	i := 0
	payload := []byte(fmt.Sprintf("hello - %v", i))
	msg := NewProduceMessage(payload)

	err = producer.Send(msg)
	if err != nil {
		panic(err)
	}

	s := msg.Get()
	log.Printf("msg send status: %v", s.String())
}

func startConsumer(c *Client) {
	consumer, err := NewConsumer(c)
	if err != nil {
		panic(err)
	}

	for {
		ctx := context.Background()
		msg, err := consumer.Receive(ctx)
		if err != nil {
			panic(err)
		}
		log.Printf("recv msg: %v", msg.String())
		err = consumer.Ack(ctx, msg)
		if err != nil {
			panic(err)
		}
	}
}
