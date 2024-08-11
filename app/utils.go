package main

import (
	"errors"
	"fmt"
	"strconv"
)

func respAsBulkString(resp string) []byte {
	var encoded string

	if respLen := len(resp); respLen == 0 {
		encoded = fmt.Sprintf("$-1%s", carriageReturn())
	} else {
		encoded = fmt.Sprintf("$%d%s%v%s", len(resp), carriageReturn(), string(resp), carriageReturn())
	}

	return []byte(encoded)
}

func respAsSimpleString(resp string) []byte {
	return []byte(fmt.Sprintf("+%s%s", resp, carriageReturn()))
}

func okSimpleString() []byte {
	return respAsSimpleString("OK")
}

func respAsArray(resp []string) ([]byte, error) {
	if len(resp) == 0 {
		return nil, errors.New("empty resp content")
	}

	encoded := make([]byte, 0, 1024)

	lenStr := strconv.Itoa(len(resp))

	encoded = append(encoded, byte(array))
	encoded = append(encoded, []byte(lenStr)...)
	encoded = append(encoded, []byte(carriageReturn())...)

	for _, r := range resp {
		encoded = append(encoded, respAsBulkString(r)...)
	}

	return encoded, nil
}

func carriageReturn() string {
	return "\r\n"
}