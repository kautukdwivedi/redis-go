package main

import (
	"strings"
)

func (s *server) handleCommandConfigGet(client *Client, args []string) error {
	key := args[0]
	var val string

	switch strings.ToLower(key) {
	case "dir":
		val = s.rdbFile.Dir
	case "dbfilename":
		val = s.rdbFile.DBFilename
	}

	if len(val) > 0 {
		resp, err := respAsArray([]string{key, val})
		if err != nil {
			return err
		}

		_, err = client.Write(resp)
		if err != nil {
			return err
		}
	}

	return nil
}
