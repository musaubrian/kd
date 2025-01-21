package main

import (
	"errors"
	"log"
	"log/slog"
	"net"
	"strings"
	"sync"
)

type DB map[string]string

type KDB struct {
	sync.RWMutex
	db DB
}

func New() *KDB {
	return &KDB{
		db: make(DB),
	}
}

func (k *KDB) Get(key string) string {
	k.RLock()
	defer k.RUnlock()

	return k.db[key]
}
func (k *KDB) Set(key, val string) {
	k.Lock()
	defer k.Unlock()
	k.db[key] = val
}

func (k *KDB) Del(key string) {
	k.Lock()
	defer k.Unlock()
	delete(k.db, key)
}

func (k *KDB) Update(key, val string) {
	k.Lock()
	defer k.Unlock()
	k.db[key] = val
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		slog.Error(err.Error())
	}
	defer ln.Close()
	for {
		con, err := ln.Accept()
		if err != nil {
			slog.Error(err.Error())
		}
		go handleRequest(con)
	}

}
func handleRequest(conn net.Conn) {
	defer conn.Close()
	k := New()
	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reading:", err)
			return
		}
		cmd := string(buf[:n])

		c, err := parseCmd(cmd)
		if err != nil {
			conn.Write([]byte(err.Error()))
			return
		}

		switch c.Operation {
		case "GET":
			v := k.Get(c.Key)
			if v == "" {
				conn.Write([]byte("-1"))
			} else {
				conn.Write([]byte(v))
			}
		case "SET":
			k.Set(c.Key, c.Val)
			conn.Write([]byte("0"))
		case "UPDATE":
			k.Update(c.Key, c.Val)
			conn.Write([]byte("0"))
		case "DEL":
			k.Del(c.Key)
			conn.Write([]byte("0"))
		default:
			conn.Write([]byte("-1"))
		}
	}
}

type Cmd struct {
	Operation string
	Key       string
	Val       string
}

func parseCmd(cmd string) (*Cmd, error) {
	cmd = strings.TrimSpace(cmd)
	splitCmds := strings.Split(cmd, " ")

	if len(splitCmds) < 2 {
		return nil, errors.New("Malformed command: at least two arguments required")
	}

	c := &Cmd{}
	op := strings.ToUpper(splitCmds[0])

	if op == "DEL" || op == "GET" {
		c = &Cmd{
			Operation: op,
			Key:       splitCmds[1],
		}
	} else if op == "SET" || op == "UPDATE" {
		if len(splitCmds) < 3 {
			return nil, errors.New("Malformed command: SET and UPDATE require both key and value")
		}
		c = &Cmd{
			Operation: op,
			Key:       splitCmds[1],
			Val:       splitCmds[2],
		}
	} else {
		return nil, errors.New("Unknown Command")
	}
	return c, nil
}
