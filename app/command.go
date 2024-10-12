package main

import (
	"strings"
)

type command struct {
	rawBytes  []byte
	name      string
	args      []string
	isQueable bool
	isWrite   bool
}

func (c *command) append(b byte) {
	c.rawBytes = append(c.rawBytes, b)
}

func (c *command) bytesLength() int {
	return len(c.rawBytes)
}

func (c *command) parse() {
	pieces := strings.Split(string(c.rawBytes), carriageReturn())[1:]

	namePieces := make([]string, 0, 2)

	for idx, piece := range pieces {
		if idx%2 == 0 || len(piece) == 0 {
			continue
		}
		if IsUpper(piece, true) {
			namePieces = append(namePieces, piece)
		} else {
			c.args = append(c.args, piece)
		}
	}

	c.name = strings.Join(namePieces, " ")
	switch c.name {
	case "ECHO", "GET", "KEYS", "TYPE":
		c.isQueable = true
	case "INCR", "SET":
		c.isQueable = true
		c.isWrite = true
	}
}
