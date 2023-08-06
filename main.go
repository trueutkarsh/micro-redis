package main

import (
	"fmt"
	"time"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func main() {

	server := microredis.NewServer("localhost", "8080", time.Second)
	client := microredis.NewClient("localhost", "8080")

	fmt.Println("Starting Server")
	go server.Run()

	time.Sleep(1 * time.Second)

	fmt.Println("Starting Client")
	client.Run()

}
