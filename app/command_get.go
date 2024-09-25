package main

import (
	"errors"
	"net"
)

func (s *server) handleCommandGet(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	s.dataMu.RLock()
	expVal, ok := s.data[args[0]]
	s.dataMu.RUnlock()
	if !ok {
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	if expVal.hasExpired() {
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := conn.Write(respAsBulkString(string(expVal.val)))
	if err != nil {
		return err
	}

	return nil
}
