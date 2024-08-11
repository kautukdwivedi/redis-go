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
	"time"
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

func (s *server) doHandshakeWithMaster() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.options.masterHost, s.options.masterPort))
	if err != nil {
		return nil
	}

	resp, err := respAsArray([]string{"PING"})
	if err != nil {
		return err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	resp, err = respAsArray([]string{"REPLCONF", "listening-port", strconv.Itoa(s.options.port)})
	if err != nil {
		return err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	resp, err = respAsArray([]string{"REPLCONF", "capa", "psync2"})
	if err != nil {
		return err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	return nil
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
