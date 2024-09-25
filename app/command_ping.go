package main

import "net"

func (s *server) handleCommandPing(conn net.Conn) error {
	if s.isMaster() {
		_, err := conn.Write(respAsSimpleString("PONG"))
		if err != nil {
			return err
		}
	}

	return nil
}
