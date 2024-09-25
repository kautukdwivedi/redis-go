package main

import (
	"net"
	"strconv"
	"strings"
)

func (s *server) handleCommandReplconfOnMaster(conn net.Conn, args []string) error {
	if strings.ToLower(args[0]) == "ack" {
		s.ackChan <- true
	} else {
		_, err := conn.Write(okSimpleString())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleCommandReplconfOnSlave(conn net.Conn, args []string) error {
	if strings.ToLower(args[0]) == "getack" && args[1] == "*" {
		resp, err := respAsArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.masterReplOffset)})
		if err != nil {
			return err
		}

		_, err = conn.Write(resp)
		if err != nil {
			return err
		}
	}

	return nil
}
