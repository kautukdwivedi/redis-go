package main

import (
	"net"
	"strconv"
)

func (s *server) handleCommandReplconf(conn net.Conn) error {
	_, err := conn.Write(okSimpleString())
	if err != nil {
		return err
	}
	return nil
}

func (s *server) handleCommandReplconfAck() error {
	s.ackChan <- true
	return nil
}

func (s *server) handleCommandReplconfGetAck(conn net.Conn) error {
	resp, err := respAsArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.masterReplOffset)})
	if err != nil {
		return err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	return nil
}
