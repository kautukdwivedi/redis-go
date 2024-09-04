package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
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
	bufStr := string(msgBuf)
	if len(bufStr) == 0 {
		return errors.New("empty raw message")
	}

	commands := strings.Split(bufStr, "*")
	if len(commands) == 0 {
		return errors.New("no commands in input message")
	}

	for _, command := range commands[1:] {
		err := s.handleCommand(conn, command)
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
	comm, err := s.findCommand(name)

	if err != nil {
		return err
	}

	comm.callback(s, conn, args)

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
