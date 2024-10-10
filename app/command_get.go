package main

import (
	"errors"
)

func (s *server) handleCommandGet(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	s.dataMu.RLock()
	expVal, ok := s.data[args[0]]
	s.dataMu.RUnlock()

	if !ok || expVal.HasExpired() {
		return nullBulkString, nil
	}

	return respAsBulkString(string(expVal.Val)), nil
}
