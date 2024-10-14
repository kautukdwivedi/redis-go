package main

import (
	"errors"
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

func (c *command) parse() error {
	pieces := strings.Split(string(c.rawBytes), carriageReturn())[1:]

	cleanCmdPieces := make([]string, 0, len(pieces)/2)

	for idx, piece := range pieces {
		if idx%2 != 0 {
			cleanCmdPieces = append(cleanCmdPieces, piece)
		}
	}

	cleanCmd := strings.ToLower(strings.Join(cleanCmdPieces, " "))

	var matchedCmd string
	for _, supportedCmd := range supportedCommands {
		if strings.HasPrefix(cleanCmd, supportedCmd) && len(matchedCmd) < len(supportedCmd) {
			matchedCmd = supportedCmd
		}
	}

	if len(matchedCmd) == 0 {
		return errors.New("command not supported")
	}

	c.name = matchedCmd

	switch c.name {
	case cmdEcho, cmdGet, cmdKeys, cmdType, cmdXRange:
		c.isQueable = true
	case cmdIncr, cmdSet, cmdXAdd:
		c.isQueable = true
		c.isWrite = true
	}

	if len(cleanCmd) > len(matchedCmd) {
		args := cleanCmd[len(matchedCmd)+1:]
		c.args = strings.Split(args, " ")
	}

	return nil
}
