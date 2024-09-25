package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"unicode"
)

type server struct {
	*serverConfig
	data     map[string]expiringValue
	dataMu   *sync.RWMutex
	slaves   []net.Conn
	slavesMu *sync.Mutex
	ackChan  chan bool
}

func newServer(config *serverConfig) server {
	return server{
		serverConfig: config,
		data:         make(map[string]expiringValue),
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

		go s.handleConn(masterConn)
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

		go s.handleConn(conn)
	}
}

func (s *server) handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading from conn: ", err)
			continue
		}

		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		s.handleRawMessage(conn, msgBuf)
	}
}

func (s *server) handleRawMessage(conn net.Conn, msgBuf []byte) error {
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
		err := s.handleCommand(conn, command)
		if err != nil {
			fmt.Println("cmd error: ", err)
		}
	}

	return nil
}

func (s *server) handleCommand(conn net.Conn, cmd *command) error {
	cmd.parse()

	if s.isMaster() {
		return s.handleCommandOnMaster(conn, cmd)
	} else {
		return s.handleCommandOnSlave(conn, cmd)
	}
}

func (s *server) handleCommandOnMaster(conn net.Conn, cmd *command) error {
	switch strings.ToLower(cmd.name) {
	case "ping":
		return s.handleCommandPing(conn)
	case "echo":
		return s.handleCommandEcho(conn, cmd.args)
	case "get":
		return s.handleCommandGet(conn, cmd.args)
	case "set":
		return s.handleCommandSetOnMaster(conn, cmd.args)
	case "info":
		return s.handleCommandInfo(conn, cmd.args)
	case "replconf":
		return s.handleCommandReplconfOnMaster(conn, cmd.args)
	case "psync":
		return s.handleCommandPsync(conn)
	case "wait":
		return s.handleCommandWait(conn, cmd.args)
	default:
		return nil
	}
}

func (s *server) handleCommandOnSlave(conn net.Conn, cmd *command) error {
	var err error

	switch strings.ToLower(cmd.name) {
	case "echo":
		err = s.handleCommandEcho(conn, cmd.args)
	case "get":
		err = s.handleCommandGet(conn, cmd.args)
	case "set":
		err = s.handleCommandSetOnSlave(cmd.args)
	case "info":
		err = s.handleCommandInfo(conn, cmd.args)
	case "replconf":
		err = s.handleCommandReplconfOnSlave(conn, cmd.args)
	}

	if err == nil {
		s.masterReplOffset += cmd.bytesLength()
	}

	return err
}

func (s *server) propagateCommandToSlaves(comm string, args []string) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)
	argsStr = append(argsStr, args...)

	resp, err := respAsArray(argsStr)
	if err != nil {
		return err
	}

	s.slavesMu.Lock()
	defer s.slavesMu.Unlock()
	for _, slave := range s.slaves {
		_, err := slave.Write(resp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}
