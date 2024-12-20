package main

import (
	"fmt"
	"net"
	"strconv"
)

func (s *server) doHandshakeWithMaster() (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.masterHost, s.masterPort))
	if err != nil {
		return nil, err
	}

	resp, err := respAsArray([]string{"PING"})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return nil, err
	}

	sleepSeconds(1)

	err = writeCommandWithArgs(conn, "REPLCONF", "listening-port", strconv.Itoa(s.port))
	if err != nil {
		return nil, err
	}

	sleepSeconds(1)

	err = writeCommandWithArgs(conn, "REPLCONF", "capa", "psync2")
	if err != nil {
		return nil, err
	}

	sleepSeconds(1)

	err = writeCommandWithArgs(conn, "PSYNC", s.masterReplId, strconv.Itoa(s.masterReplOffset))
	if err != nil {
		return nil, err
	}

	s.masterReplOffset = 0

	sleepSeconds(1)

	return conn, nil
}

func writeCommandWithArgs(conn net.Conn, command string, args ...string) error {
	resp := make([]string, 0, 1+len(args))

	resp = append(resp, command)
	resp = append(resp, args...)

	respArr, err := respAsArray(resp)
	if err != nil {
		return err
	}

	_, err = conn.Write(respArr)
	if err != nil {
		return err
	}

	return nil
}
