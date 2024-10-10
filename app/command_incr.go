package main

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func (s *server) handleCommandIncr(args []string) ([]byte, error) {
	key := args[0]

	s.dataMu.RLock()
	expVal, ok := s.data[key]
	s.dataMu.RUnlock()
	var newVal int
	if !ok {
		newVal = 1
		s.dataMu.Lock()
		s.data[key] = storage.NewExpiringValue(strconv.Itoa(newVal))
		s.dataMu.Unlock()
	} else {
		val, err := strconv.Atoi(expVal.Val)
		if err != nil {
			return respAsError("value is not an integer or out of range"), nil
		} else {
			newVal = val + 1
			s.dataMu.Lock()
			s.data[key].Val = strconv.Itoa(newVal)
			s.dataMu.Unlock()
		}
	}

	return respAsInteger(newVal), nil
}
