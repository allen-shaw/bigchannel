package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	sendMsgs []string
	recvMsgs []string
)

func main() {
	addr := "127.0.0.1:18088"
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}
	log.Println("new client succ")

	/*go*/
	startProducer(client)
	go startConsumer(client)

	time.Sleep(5 * time.Second)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-sig

	fmt.Println("send msgs:")
	fmt.Println(sendMsgs)
	fmt.Println("recv msgs:")
	fmt.Println(recvMsgs)
}

func startProducer(c *Client) {
	producer, err := NewProducer(c)
	if err != nil {
		panic(err)
	}
	log.Println("new producer success")
	now := time.Now().Format("15:04:05")

	for i := 0; i < 10; i++ {
		payload := []byte(fmt.Sprintf("hello-%v-%v", now, i))
		msg := NewProduceMessage(payload)

		err = producer.Send(msg)
		if err != nil {
			panic(err)
		}
		sendMsgs = append(sendMsgs, string(msg.m.Payload))

		s := msg.Get()
		if s != StatusSucc {
			log.Fatalf("invalid status %v: %v", s.String(), s)
		}
		log.Println("producer send msg success", i)
		time.Sleep(1 * time.Second)
	}

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
		if msg == nil {
			log.Printf("recv nil: %v", msg)
			panic("msg is nil")
		}
		log.Printf("recv msg: %v", string(msg.Payload))
		recvMsgs = append(recvMsgs, string(msg.Payload))
		err = consumer.Ack(ctx, msg)
		if err != nil {
			panic(err)
		}
	}
}
