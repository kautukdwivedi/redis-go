package main

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func (s *server) handleCommandIncr(client *Client, args []string) error {
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
			_, err = client.Write(respAsError("value is not an integer or out of range"))
			if err != nil {
				return err
			}
			return nil
		} else {
			newVal = val + 1
			s.data[key].Val = strconv.Itoa(newVal)
		}
	}
	s.dataMu.Unlock()

	_, err := client.Write(respAsInteger(newVal))
	if err != nil {
		return err
	}

	return nil
}
