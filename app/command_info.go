package main

import (
	"errors"
	"strings"
)

func (s *server) handleCommandInfo(client *Client, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")
		return respAsBulkString(respStr), nil
	}

	return nil, nil
}
