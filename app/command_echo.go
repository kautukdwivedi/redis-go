package main

import (
	"errors"
	"net"
)

func (s *server) handleCommandEcho(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	_, err := conn.Write(respAsSimpleString(args[0]))
	if err != nil {
		return err
	}

	return nil
}
