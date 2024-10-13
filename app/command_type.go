package main

import (
	"fmt"
	"reflect"
)

func (s *server) handleCommandType(args []string) ([]byte, error) {
	val, ok := s.data[args[0]]
	if !ok {
		stream := s.findStream(args[0])
		if stream == nil {
			return respAsSimpleString("none"), nil
		}
		return respAsSimpleString("stream"), nil
	}

	t := reflect.TypeOf(val.Val).Kind()
	if t == reflect.String {
		return respAsSimpleString("string"), nil
	}

	return nil, fmt.Errorf("value type not supported: %v", t)
}
