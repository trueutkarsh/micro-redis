package main

import (
	"fmt"
	"time"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func main() {

	server := microredis.NewServer("localhost", "6379", time.Second)
	client := microredis.NewClient("localhost", "6379")

	fmt.Println("Starting Server")
	go server.Run()

	time.Sleep(1 * time.Second)

	fmt.Println("Starting Client")
	client.Run()

}
