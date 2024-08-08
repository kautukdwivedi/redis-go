package main

import "fmt"

func respAsBulkString(resp string) []byte {
	var encoded string

	if respLen := len(resp); respLen == 0 {
		encoded = "$-1\r\n"
	} else {
		encoded = fmt.Sprintf("$%d\r\n%v\r\n", len(resp), string(resp))
	}

	return []byte(encoded)
}

func respAsSimpleString(resp string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", resp))
}

func okSimpleString() []byte {
	return respAsSimpleString("OK")
}
