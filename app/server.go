package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
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
}

func newServer(config *serverConfig) server {
	return server{
		serverConfig: config,
		data:         make(map[string]expiringValue),
		dataMu:       &sync.RWMutex{},
		slaves:       []net.Conn{},
		slavesMu:     &sync.Mutex{},
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
	startIdx := -1
	count := 0
	commands := [][]byte{}
outer:
	for idx, b := range msgBuf {
		bRune := rune(b)
		switch {
		case bRune == '*':
			startIdx = idx
		case unicode.IsDigit(bRune) && startIdx > -1:
			if count < 10 {
				commands = append(commands, make([]byte, 0))
			}
			val, err := strconv.Atoi(string(bRune))
			if err != nil {
				break outer
			}
			count = count*10 + val
		default:
			if len(commands) == 0 {
				commands = append(commands, make([]byte, 0))
			}
			lastIdx := len(commands) - 1
			if startIdx >= 0 && count == 0 {
				commands[lastIdx] = append(commands[lastIdx], '*')
			}
			commands[lastIdx] = append(commands[lastIdx], b)
			startIdx = -1
			count = 0
		}
	}

	for _, command := range commands {
		err := s.handleCommand(conn, string(command))
		if err != nil {
			fmt.Println("cmd error: ", err)
		}

	}

	return nil
}

func (s *server) handleCommand(conn net.Conn, cmd string) error {
	cmdPieces := strings.Split(cmd, carriageReturn())

	if len(cmdPieces) <= 1 {
		return errors.New("command string is not a valid command")
	}

	name, args := parseCommand(cmdPieces[1:])

	switch strings.ToLower(name) {
	case "ping":
		s.handleCommandPing(conn)
	case "echo":
		s.handleCommandEcho(conn, args)
	case "get":
		s.handleCommandGet(conn, args)
	case "set":
		s.handleCommandSet(conn, args)
	case "info":
		s.handleCommandInfo(conn, args)
	case "replconf":
		s.handleCommandReplconf(conn, args)
	case "psync":
		s.handleCommandPsync(conn)
	default:
		return fmt.Errorf("unknown command: \"%s\"", name)
	}

	return nil
}

func (s *server) propagateCommandToSlaves(comm string, args []string) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)
	argsStr = append(argsStr, args...)

	resp, err := respAsArray(argsStr)
	if err != nil {
		return err
	}

	fmt.Printf("Propagating commands to %d slaves\n", len(s.slaves))

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
