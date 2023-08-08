package microredis

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
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
	listener, err := net.Listen("tcp", s.address+":"+s.port)
	if err != nil {
		log.Fatal(err)
	}

	// background expired keys clean up goroutine
	go func(s *Server) {
		for {
			s.lock.Lock()
			time.Sleep(s.db.clear_freq)
			s.db.ClearExpiredKeys()
			s.lock.Unlock()
		}
	}(s)

	// listen for messages
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "quit" {
			return
		}

		if line == "" || line == "\n" {
			continue
		}

		var result string
		response, err := s.processRESP(line)

		if err != nil {
			result = MarshalResp(err)
		} else {
			result = MarshalResp(response)
		}
		if _, err := conn.Write([]byte(result + "\n")); err != nil {
			fmt.Printf("Failed to write response: %v\n", err)
			continue
		}
	}
}

func (s *Server) processRESP(msg string) (interface{}, error) {
	s.lock.Lock()         // aquire lock
	defer s.lock.Unlock() // release lock when processing done
	commands, err := UnmarshalResp(msg)
	if err != nil {
		return "", err
	}
	switch commands[0] {
	case "GET":
		return s.ProcessRespCommandGet(commands)

	case "SET":
		return s.ProcessRespCommandSet(commands)

	case "DEL":
		return s.ProcessRespCommandDel(commands)

	case "EXPIRE":
		return s.ProcessRespCommandExpire(commands)

	case "TTL":
		return s.ProcessRespCommandTTL(commands)

	case "KEYS":
		return s.ProcessRespCommandKeys(commands)

	default:
		return nil, errors.New(fmt.Sprintf("Invalid command %s", commands[0]))
	}

}

func (s *Server) ProcessRespCommandGet(commands []string) (interface{}, error) {
	if len(commands) > 2 {
		return nil, errors.New("ERR Wrong number of arguments")
	}
	result := s.db.Get(Key(commands[1]))
	if result == nil {
		return nil, nil
	} else {
		return *result, nil
	}
}

func (s *Server) ProcessRespCommandSet(commands []string) (interface{}, error) {
	if len(commands) > 7 || len(commands) < 3 {
		return nil, errors.New("ERR Wrong number of commands")
	}

	key := Key(commands[1])
	val := commands[2]
	var exp time.Time
	ret_old_val := false
	keep_ttl := false
	set_if_exists := false
	set_if_not_exists := false
	expiry_set := false

	args := make(map[string]string)

	i := 3
	for {
		if i == len(commands) {
			break
		}
		if strings.ToUpper(commands[i]) == "NX" {
			args["NX"] = "true"
			if _, prs := args["XX"]; prs {
				return nil, errors.New("ERR Invalid args NX and XX can't be present together")
			}
			set_if_not_exists = true
			i += 1
		} else if strings.ToUpper(commands[i]) == "XX" {
			args["XX"] = "true"
			if _, prs := args["NX"]; prs {
				return nil, errors.New("ERR Invalid args NX and XX can't be present together")
			}
			set_if_exists = true
			i += 1
		} else if strings.ToUpper(commands[i]) == "GET" {
			ret_old_val = true
			i += 1
		} else if strings.ToUpper(commands[i]) == "EX" {
			if !expiry_set {
				sec, err := strconv.ParseFloat(commands[i+1], 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("ERR Invalid args, unable to parse %s", commands[i+1]))
				}
				exp = time.Now().Add(time.Duration(float64(sec) * float64(time.Second)))
				i += 2
				expiry_set = true
			} else {
				return nil, errors.New("Invalid Args, multiple expiry provided")
			}
		} else if strings.ToUpper(commands[i]) == "PX" {
			if !expiry_set {
				milsec, err := strconv.ParseInt(commands[i+1], 10, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("ERR Invalid args, unable to parse %s", commands[i+1]))
				}
				exp = time.Now().Add(time.Duration(float64(milsec) * float64(time.Millisecond)))
				i += 2
				expiry_set = true
			} else {
				return nil, errors.New("Invalid Args, multiple expiry provided")
			}
		} else if strings.ToUpper(commands[i]) == "EXAT" {
			if !expiry_set {
				sec, err := strconv.ParseInt(commands[i+1], 10, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("ERR Invalid args, unable to parse %s", commands[i+1]))
				}
				exp = time.Unix(sec, 0)
				i += 2
				expiry_set = true
			} else {
				return nil, errors.New("Invalid Args, multiple expiry provided")
			}
		} else if strings.ToUpper(commands[i]) == "PXAT" {
			if !expiry_set {
				milsec, err := strconv.ParseInt(commands[i+1], 10, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("ERR Invalid args, unable to parse %s", commands[i+1]))
				}
				exp = time.Unix(0, milsec*1000)
				i += 2
				expiry_set = true
			} else {
				return nil, errors.New("Invalid Args, multiple expiry provided")
			}
		} else if strings.ToUpper(commands[i]) == "KEEPTTL" {
			if !expiry_set {
				// exp = nil
				i += 1
				expiry_set = true
			} else {
				return nil, errors.New("Invalid Args, multiple expiry provided")
			}
		} else {
			return nil, errors.New(fmt.Sprintf("Invalid Arg: %s", commands[i]))
		}

	}

	var success bool
	var old_val *string
	if keep_ttl || !expiry_set {
		success, old_val = s.db.Set(key, val, nil, ret_old_val, keep_ttl, set_if_exists, set_if_not_exists)
	} else {
		success, old_val = s.db.Set(key, val, &exp, ret_old_val, keep_ttl, set_if_exists, set_if_not_exists)
	}
	if !success {
		return nil, nil
	} else {
		if !ret_old_val {
			return "OK", nil
		} else {
			if old_val == nil {
				return nil, nil
			} else {
				return *old_val, nil
			}
		}
	}
}

func (s *Server) ProcessRespCommandDel(commands []string) (interface{}, error) {
	if len(commands) < 2 {
		return nil, errors.New("ERR Invalid number of args")
	}
	keys := make([]Key, 0)
	for _, c := range commands[1:] {
		keys = append(keys, Key(c))
	}
	return s.db.Del(keys), nil
}

func (s *Server) ProcessRespCommandExpire(commands []string) (interface{}, error) {
	if len(commands) < 3 || len(commands) > 4 {
		return nil, errors.New("ERR invalid number of args")
	}

	key := Key(commands[1])
	secs, err := strconv.ParseInt(commands[2], 10, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERR Invalid arg: Unable to parse sec argument %s", commands[2]))
	}
	set_if_no_expiry := false
	set_if_expiry := false
	set_if_gt := false
	set_if_lt := false

	if len(commands) == 4 {
		if commands[3] == "NX" {
			set_if_no_expiry = true
		} else if commands[3] == "XX" {
			set_if_expiry = true
		} else if commands[3] == "GT" {
			set_if_gt = true
		} else if commands[3] == "LT" {
			set_if_lt = true
		}
	}

	return s.db.Expire(key, secs, set_if_no_expiry, set_if_expiry, set_if_gt, set_if_lt), nil
}

func (s *Server) ProcessRespCommandTTL(commands []string) (interface{}, error) {
	if len(commands) != 2 {
		return nil, errors.New("ERR Invalid number of args")
	}
	result := s.db.TTL(Key(commands[1]))
	return result, nil
}

func (s *Server) ProcessRespCommandKeys(commands []string) (interface{}, error) {
	if len(commands) != 2 {
		return nil, errors.New("ERR Invalid number of args")
	}

	result, err := s.db.Keys(commands[1])
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
