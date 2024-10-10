package main

import (
	"errors"
)

func (s *server) handleCommandEcho(client *Client, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("command echo must take one argument")
	}

	return respAsSimpleString(args[0]), nil
}
