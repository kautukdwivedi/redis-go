package main

import (
	"fmt"
	"reflect"
)

func (s *server) handleCommandType(args []string) ([]byte, error) {
	val, ok := s.data[args[0]]
	if !ok {
		return respAsSimpleString("none"), nil
	}

	t := reflect.TypeOf(val.Val).Kind()
	if t == reflect.String {
		return respAsSimpleString("string"), nil
	}

	return nil, fmt.Errorf("value type not supported: %v", t)
}
