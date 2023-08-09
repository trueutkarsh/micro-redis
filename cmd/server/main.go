package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func main() {

	addressPtr := flag.String("address", "localhost", "address to connect to")
	portPtr := flag.String("port", "6379", "port to connect to")
	clearFreqPtr := flag.Int64(
		"clearfreq",
		1000,
		"millseconds after which to clear expired keys in storage",
	)

	flag.Parse()

	server := microredis.NewServer(*addressPtr, *portPtr, time.Duration(*clearFreqPtr*int64(time.Millisecond)))

	fmt.Printf("Starting Server at %s:%s \n", *addressPtr, *portPtr)
	server.Run()

}
