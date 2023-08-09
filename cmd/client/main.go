package main

import (
	"flag"
	"fmt"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func main() {

	addressPtr := flag.String("address", "localhost", "address to connect to")
	portPtr := flag.String("port", "6379", "port to connect to")

	flag.Parse()

	client := microredis.NewClient(*addressPtr, *portPtr)

	fmt.Println("Starting Client")
	client.Run()

}
