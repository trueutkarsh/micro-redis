package microredis

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
		log.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	conn_scanner := bufio.NewScanner(conn)
	fmt.Print(fmt.Sprintf("redis:%s:%s> ", c.address, c.port))
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
		fmt.Print(fmt.Sprintf("redis:%s:%s> ", c.address, c.port))
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from terminal:", err)
	}
}
