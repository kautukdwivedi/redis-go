package main

import "net"

func (s *server) handleCommandMulti(conn net.Conn) error {
	_, err := conn.Write(okSimpleString())
	if err != nil {
		return err
	}
	return nil
}
