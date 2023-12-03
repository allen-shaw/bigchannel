package main

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

func TestProduceConsumeSameTime(t *testing.T) {
	addr := "127.0.0.1:18088"
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}
	log.Println("new client succ")

	/*go*/
	go startProducer(client)
	go startConsumer(client)

	time.Sleep(25 * time.Second)

	fmt.Println("send msgs:")
	fmt.Println(strings.Join(sendMsgs, "; "))
	fmt.Println("recv msgs:")
	fmt.Println(strings.Join(recvMsgs, "; "))
}

func TestProduceFirst(t *testing.T) {
	addr := "127.0.0.1:18088"
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}
	log.Println("new client succ")

	/*go*/
	startProducer(client)
	go startConsumer(client)

	time.Sleep(10 * time.Second)

	fmt.Println("send msgs:")
	fmt.Println(strings.Join(sendMsgs, "; "))
	fmt.Println("recv msgs:")
	fmt.Println(strings.Join(recvMsgs, "; "))
}

func TestProduceFirstThenConsumeThenProduceAgain(t *testing.T) {

}

func TestConsumeFirst(t *testing.T) {
	addr := "127.0.0.1:18088"
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}
	log.Println("new client succ")

	/*go*/
	go startConsumer(client)
	time.Sleep(3 * time.Second)
	startProducer(client)

	time.Sleep(10 * time.Second)

	fmt.Println("send msgs:")
	fmt.Println(strings.Join(sendMsgs, "; "))
	fmt.Println("recv msgs:")
	fmt.Println(strings.Join(recvMsgs, "; "))
}

func TestAck(t *testing.T) {

}

func TestRestartNoAck(t *testing.T) {

}
