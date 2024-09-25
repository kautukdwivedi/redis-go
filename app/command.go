package main

import (
	"fmt"
	"strings"
)

type commands struct {
	data []*command
}

func (cmds *commands) append(cmd *command) {
	cmds.data = append(cmds.data, cmd)
}

type command struct {
	rawBytes []byte
	name     string
	args     []string
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
	fmt.Println("Cmd name: ", c.name)
	fmt.Println("Cmd args: ", c.args)
}
