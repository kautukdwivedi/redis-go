package main

import (
	"strings"
)

func (s *server) handleCommandConfigGet(client *Client, args []string) ([]byte, error) {
	key := args[0]
	var val string

	switch strings.ToLower(key) {
	case "dir":
		val = s.rdbFile.Dir
	case "dbfilename":
		val = s.rdbFile.DBFilename
	}

	if len(val) > 0 {
		return respAsArray([]string{key, val})
	}

	return nil, nil
}
