package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"unicode"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

type server struct {
	*serverConfig
	data         map[string]*storage.ExpiringValue
	dataMu       *sync.RWMutex
	slaves       []net.Conn
	slavesMu     *sync.Mutex
	ackChan      chan bool
}

func newServer(config *serverConfig) server {
	return server{
		serverConfig: config,
		data:         make(map[string]*storage.ExpiringValue),
		dataMu:       &sync.RWMutex{},
		slaves:       []net.Conn{},
		slavesMu:     &sync.Mutex{},
		ackChan:      make(chan bool),
	}
}

func (s *server) isMaster() bool {
	return s.role == master
}

func (s *server) isSlave() bool {
	return s.role == slave
}

func (s *server) start() {
	if s.isSlave() {
		masterConn, err := s.doHandshakeWithMaster()
		if err != nil {
			fmt.Println("Error connecting to master: ", err)
			return
		}

		go s.handleClient(masterConn)
	}

	log.Fatal(s.listenAndServe())
}

func (s *server) listenAndServe() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		go s.handleClient(conn)
	}
}

func (s *server) handleClient(conn net.Conn) {
	client := NewClient(conn)
	defer client.Close()

	buf := make([]byte, 1024)
	for {
		n, err := client.Conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading from conn: ", err)
			continue
		}

		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		s.handleRawMessage(client, msgBuf)
	}
}

func (s *server) handleRawMessage(client *Client, msgBuf []byte) error {
	cmds := &commands{
		data: make([]*command, 0),
	}
	var cmd *command

	for _, b := range msgBuf {
		bRune := rune(b)
		switch {
		case bRune == '*':
			cmd = &command{
				rawBytes: []byte{b},
				args:     make([]string, 0),
			}
			cmds.append(cmd)
		case unicode.IsDigit(bRune):
			if cmd == nil {
				continue
			}
			cmd.append(b)
		default:
			if cmd == nil {
				continue
			}
			n := len(cmds.data) - 1
			lastCmd := cmds.data[n]
			if len(lastCmd.rawBytes) == 1 && lastCmd.rawBytes[0] == '*' {
				cmds.data = cmds.data[:n]
				n = len(cmds.data) - 1
				cmd = cmds.data[n]
				cmd.append('*')
			}
			cmd.append(b)
		}
	}

	for _, command := range cmds.data {
		err := s.handleCommand(client, command)
		if err != nil {
			fmt.Println("cmd error: ", err)
		}
	}

	return nil
}

func (s *server) loadRDB() (bool, error) {
	err := s.rdbFile.Load()
	if err != nil {
		return false, err
	}

	s.dataMu.Lock()
	now := time.Now().UTC()
	for _, db := range s.rdbFile.DBs {
		for key, rdbObj := range db.Data {
			var expiresIn int
			if rdbObj.Exp != nil {
				expiresIn = int(rdbObj.Exp.Sub(now).Milliseconds())
			}
			s.data[key] = &storage.ExpiringValue{
				Val:       rdbObj.Val,
				Created:   now,
				ExpiresIn: expiresIn,
			}
		}
	}
	s.dataMu.Unlock()

	return true, nil
}

func (s *server) getKeys() []string {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}
