package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	port := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "<MASTER_HOST> <MASTER_PORT>")
	flag.Parse()

	s := &server{
		options: &serverOptions{},
	}

	err := s.parseReplicaOf(*replicaOf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if s.isSlave() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.options.masterHost, s.options.masterPort))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		fmt.Println("Failed to bind to port ", *port)
		os.Exit(1)
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

func (s *server) parseReplicaOf(replicaOf string) error {
	if len(replicaOf) > 0 {
		addrAndPort := strings.Split(replicaOf, " ")

		if len(addrAndPort) != 2 {
			return errors.New("invalid replica address format")
		}

		port, err := strconv.Atoi(addrAndPort[1])
		if err != nil {
			return errors.New("invalid replica port")
		}

		s.options.role = slave
		s.options.masterHost = addrAndPort[0]
		s.options.masterPort = port
		s.options.masterReplOffset = -1
	} else {
		s.options.role = master
		s.options.masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.options.masterReplOffset = 0
	}

	return nil
}

func (s *server) handleClient(conn net.Conn, client *client) {
	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			return
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
