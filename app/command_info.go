package main

import (
	"errors"
	"strings"
)

func (s *server) handleCommandInfo(client *Client, args []string) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")
		_, err := client.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}
