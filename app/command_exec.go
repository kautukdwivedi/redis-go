package main

import "net"

func (s *server) handleCommandExec(conn net.Conn) error {
	_, err := conn.Write(respAsError("EXEC without MULTI"))
	if err != nil {
		return err
	}
	return nil
}
