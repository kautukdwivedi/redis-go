package main

import "strings"

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
	pieces := strings.Split(string(c.rawBytes), carriageReturn())[2:]
	c.name = pieces[0]
	if len(pieces) > 1 {
		rawArgs := pieces[1:]
		c.args = make([]string, 0, len(rawArgs)/2)
		for idx, piece := range rawArgs {
			if idx%2 != 0 {
				c.args = append(c.args, piece)
			}
		}
	}
}
