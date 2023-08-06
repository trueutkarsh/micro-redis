package main

import (
	"time"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func main() {

	server := microredis.NewServer("localhost", "6969", time.Second)
	client := microredis.NewClient("localhost", "6969")
	go server.Run()

	client.Run()

}
