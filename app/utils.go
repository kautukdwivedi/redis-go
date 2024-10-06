package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
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

func respAsFileData(data []byte) []byte {
	resp := []byte{}

	resp = append(resp, byte(bulkString))
	resp = append(resp, []byte(strconv.Itoa(len(data)))...)
	resp = append(resp, []byte(carriageReturn())...)
	resp = append(resp, data...)

	return resp
}

func respAsInteger(data int) []byte {
	var sb strings.Builder
	sb.WriteString(string(integer))
	sb.WriteString(strconv.Itoa(data))
	sb.WriteString(carriageReturn())
	return []byte(sb.String())
}

func respAsError(err string) []byte {
	errStr := fmt.Sprint("-ERR", " ", err, carriageReturn())
	return []byte(errStr)
}

func carriageReturn() string {
	return "\r\n"
}

func sleepSeconds(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

func intToByteSlice(input int) ([]byte, error) {
	num := int64(input)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func IsUpper(s string, onlyLetters bool) bool {
	for _, r := range s {
		if onlyLetters && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
