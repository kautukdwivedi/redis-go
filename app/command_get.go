package main

import (
	"errors"
	"fmt"
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
	fmt.Printf("result for getting \"%s\" is \"%v\"", args[0], ok)
	if !ok {
		fmt.Println("Writing null")
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	if expVal.HasExpired() {
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := conn.Write(respAsBulkString(string(expVal.Val)))
	if err != nil {
		return err
	}

	return nil
}
