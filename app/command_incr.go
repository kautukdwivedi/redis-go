package main

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func (s *server) handleCommandIncr(client *Client, args []string) ([]byte, error) {
	key := args[0]

	s.dataMu.Lock()
	expVal, ok := s.data[key]
	var newVal int
	if !ok {
		newVal = 1
		s.data[key] = storage.NewExpiringValue(strconv.Itoa(newVal))
	} else {
		val, err := strconv.Atoi(expVal.Val)
		if err != nil {
			return respAsError("value is not an integer or out of range"), nil
		} else {
			newVal = val + 1
			s.data[key].Val = strconv.Itoa(newVal)
		}
	}
	s.dataMu.Unlock()

	return respAsInteger(newVal), nil
}
