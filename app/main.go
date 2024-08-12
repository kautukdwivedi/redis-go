package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	port := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "<MASTER_HOST> <MASTER_PORT>")
	flag.Parse()

	if port == nil {
		fmt.Println("Server port is nil")
		os.Exit(1)
	}

	s := &server{
		options: &serverOptions{
			port: *port,
		},
	}

	err := s.parseReplicaOf(*replicaOf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if s.isSlave() {
		err = s.doHandshakeWithMaster()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	err = s.listenAndServe()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (s *server) listenAndServe() error {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.options.port))
	if err != nil {
		return fmt.Errorf("failed to bind to port %v", s.options.port)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		client := &client{
			mapData: map[string]expiringValue{},
		}

		go s.handleClient(conn, client)
	}

}

func (s *server) handleClient(conn net.Conn, client *client) {
	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			continue
		}

		fmt.Println("Values read: ", strings.Split(string(buf), "\r\n"))

		err = s.parseInputBuffer(buf, conn, client)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (s *server) parseInputBuffer(buf []byte, conn net.Conn, client *client) error {
	if len(buf) == 0 {
		return errors.New("empty input buffer")
	}

	if len(buf) == 1 {
		return errors.New("input only contains data type information, but no data")
	}

	var t Type = Type(buf[0])
	switch t {
	case array:
		splitBuf := bytes.Split(buf, []byte("\r\n"))

		if len(splitBuf) == 1 {
			return errors.New("input data is an array, but does not contain actual data")
		}

		name, args := parseCommand(splitBuf[1:])

		comm, err := s.findCommand(name)
		if err != nil {
			return err
		}

		comm.callback(conn, client, args)
	}
	return nil
}
