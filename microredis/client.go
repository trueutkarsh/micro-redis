package microredis

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Client struct denotes redis client which creates a tcp
// connection to redis server. Address and port are the
// ip address to which the client will connect
type Client struct {
	address string
	port    string
}

// NewClient function creates and initializes new Redis client
// instance
func NewClient(address string, port string) *Client {
	result := Client{
		address: address,
		port:    port,
	}
	return &result
}

// Run is the function that dials a tcp connection to the given address
// and continuosly reads from the terminal (stdin) and on every new line
// line character or press enter, it marshals the message into appropriate
// RESP format and sends it server. Next it waits for the response from the
// server, unmarshals RESP msg and prints the result on terminal
func (c *Client) Run() {
	conn, err := net.Dial("tcp", c.address+":"+c.port)
	if err != nil {
		fmt.Println("Error connecting to server", err)
		log.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	conn_scanner := bufio.NewScanner(conn)
	fmt.Print(fmt.Sprintf("redis://%s:%s> ", c.address, c.port))
	for scanner.Scan() {
		msg := scanner.Text()
		msg = strings.Trim(msg, " ")
		if msg == "quit" {
			_, err := conn.Write([]byte("quit" + "\n"))
			if err != nil {
				fmt.Println("Error writing to server:", err)
			}
			return
		}
		marshalled_msg := MarshalResp(strings.Split(msg, " "))
		_, err := conn.Write([]byte(marshalled_msg + "\n"))
		if err != nil {
			fmt.Println("Error writing to server:", err)
			continue
		}
		// read from server
		line := ""
		for conn_scanner.Scan() {
			line = conn_scanner.Text()
			break
		}

		response, err := UnmarshalResp(line)
		if err != nil {
			fmt.Println("Err:-", err)
		} else {
			fmt.Println(response)
		}
		fmt.Print(fmt.Sprintf("redis://%s:%s> ", c.address, c.port))
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from terminal:", err)
	}
}
