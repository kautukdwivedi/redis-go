package main

import (
	"errors"
)

func (s *server) handleCommandEcho(client *Client, args []string) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	_, err := client.Write(respAsSimpleString(args[0]))
	if err != nil {
		return err
	}

	return nil
}
