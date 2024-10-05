package main

import "net"

func (s *server) handleCommandKeys(conn net.Conn) error {
	resp, err := respAsArray(s.getKeys())
	if err != nil {
		return err
	}

	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	return nil
}
