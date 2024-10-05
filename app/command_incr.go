package main

import (
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func (s *server) handleCommandIncr(conn net.Conn, args []string) error {
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
			return err
		}
		newVal = val + 1
		s.data[key].Val = strconv.Itoa(newVal)
	}
	s.dataMu.Unlock()

	_, err := conn.Write(respAsInteger(newVal))
	if err != nil {
		return err
	}

	return nil
}
