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

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
)

type server struct {
	role             string
	masterReplId     string
	masterReplOffset *int
}

func (s *server) isMaster() bool {
	return strings.EqualFold(s.role, "master")
}

func main() {
	port := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "Server, whose slave this instance is.")
	flag.Parse()

	s := &server{}

	err := s.parseReplicaOf(*replicaOf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
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

		if _, err := strconv.Atoi(addrAndPort[1]); err != nil {
			return errors.New("invalid replica port")
		}

		s.role = "slave"
	} else {
		s.role = "master"
		s.masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		offset := 0
		s.masterReplOffset = &offset
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
