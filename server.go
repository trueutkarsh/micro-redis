package microredis

import (
	"log"
	"net"
	"os"
	"sync"
	"time"
	//"github.com/mediocregopher/radix/v4/resp/resp3"
)

type Server struct {
	db      *Storage
	address string
	port    string
	lock    *sync.Mutex
}

func NewServer(address string, port string, clear_freq time.Duration) *Server {
	result := Server{
		db:      NewStorage(clear_freq),
		address: address,
		port:    port,
		lock:    &sync.Mutex{},
	}
	return &result
}

func (s *Server) Run() {
	// open a tcp connection
	listen, err := net.Listen("tcp", s.address+":"+s.port)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// listen for messages
	// parse them as array messages
}
