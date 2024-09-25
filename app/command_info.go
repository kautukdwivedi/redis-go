package main

import (
	"errors"
	"net"
	"strings"
)

func (s *server) handleCommandInfo(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")

		_, err := conn.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}
