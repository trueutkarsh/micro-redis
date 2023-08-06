package microredis

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Client struct {
	address string
	port    string
}

func NewClient(address string, port string) *Client {
	result := Client{
		address: address,
		port:    port,
	}
	return &result
}

func (c *Client) Run() {
	conn, err := net.Dial("tcp", c.address+":"+c.port)
	if err != nil {
		fmt.Println("Error connecting to server", err)
	}
	defer conn.Close()

	// Start a goroutine to handle incoming messages from the server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println("Received from server:", scanner.Text())
			response, err := UnmarshalResp(scanner.Text())
			if err != nil {
				fmt.Println("Err: ", err)
			} else {
				fmt.Println(response)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from server:", err)
		}
	}()

	// In the main goroutine, read lines from the terminal, split and marshal them and send them to the server
	scanner := bufio.NewScanner(os.Stdin)
	conn_scanner := bufio.NewScanner(conn)
	fmt.Print(fmt.Sprintf("redis:%s:%s> ", c.address, c.port))
	for scanner.Scan() {
		msg := scanner.Text()
		if msg == "quit" {
			return
		}
		fmt.Println("Input: received this from terminal", msg)
		marshalled_msg := MarshalResp(strings.Split(msg, " "))
		fmt.Print("sending this to server", marshalled_msg, conn.LocalAddr())
		_, err := conn.Write([]byte(marshalled_msg))
		if err != nil {
			fmt.Println("Error writing to server:", err)
			break
		}
		// read from server
		line := conn_scanner.Text()
		for conn_scanner.Scan() {
			line += conn_scanner.Text()
		}

		if err := conn_scanner.Err(); err != nil {
			fmt.Printf("Failed to read from connection: %v", err)
		}

		fmt.Println("Received from server:", line)
		response, err := UnmarshalResp(scanner.Text())
		if err != nil {
			fmt.Println("Err: ", err)
		} else {
			fmt.Println(response)
		}
		time.Sleep(100 * time.Millisecond)
		fmt.Print(fmt.Sprintf("redis:%s:%s> ", c.address, c.port))
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from terminal:", err)
	}
}
