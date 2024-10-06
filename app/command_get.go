package main

import (
	"errors"
)

func (s *server) handleCommandGet(client *Client, args []string) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	s.dataMu.RLock()
	expVal, ok := s.data[args[0]]
	s.dataMu.RUnlock()
	if !ok {
		_, err := client.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	if expVal.HasExpired() {
		_, err := client.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := client.Write(respAsBulkString(string(expVal.Val)))
	if err != nil {
		return err
	}

	return nil
}
