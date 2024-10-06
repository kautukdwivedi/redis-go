package main

import (
	"fmt"
	"strings"
)

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
	integer    Type = ':'
)

func (s *server) handleCommand(client *Client, cmd *command) error {
	cmd.parse()

	if s.isMaster() {
		return s.handleCommandOnMaster(client, cmd)
	} else {
		return s.handleCommandOnSlave(client, cmd)
	}
}

func (s *server) handleCommandOnMaster(client *Client, cmd *command) error {
	switch strings.ToLower(cmd.name) {
	case "ping":
		return s.handleCommandPing(client)
	case "echo":
		return s.handleCommandEcho(client, cmd.args)
	case "get":
		return s.handleCommandGet(client, cmd.args)
	case "set":
		return s.handleCommandSetOnMaster(client, cmd.args)
	case "info":
		return s.handleCommandInfo(client, cmd.args)
	case "replconf":
		return s.handleCommandReplconf(client)
	case "replconf ack":
		return s.handleCommandReplconfAck()
	case "psync":
		return s.handleCommandPsync(client)
	case "wait":
		return s.handleCommandWait(client, cmd.args)
	case "config get":
		return s.handleCommandConfigGet(client, cmd.args)
	case "keys":
		return s.handleCommandKeys(client)
	case "incr":
		return s.handleCommandIncr(client, cmd.args)
	case "multi":
		return s.handleCommandMulti(client)
	case "exec":
		return s.handleCommandExec(client)
	default:
		return nil
	}
}

func (s *server) handleCommandOnSlave(client *Client, cmd *command) error {
	var err error

	switch strings.ToLower(cmd.name) {
	case "echo":
		err = s.handleCommandEcho(client, cmd.args)
	case "get":
		err = s.handleCommandGet(client, cmd.args)
	case "set":
		err = s.handleCommandSetOnSlave(cmd.args)
	case "info":
		err = s.handleCommandInfo(client, cmd.args)
	case "replconf getack":
		err = s.handleCommandReplconfGetAck(client)
	case "config get":
		err = s.handleCommandConfigGet(client, cmd.args)
	case "keys":
		err = s.handleCommandKeys(client)
	case "incr":
		err = s.handleCommandIncr(client, cmd.args)
	}

	if err == nil {
		s.masterReplOffset += cmd.bytesLength()
	}

	return err
}

func (s *server) propagateCommandToSlaves(comm string, args []string) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)
	argsStr = append(argsStr, args...)

	resp, err := respAsArray(argsStr)
	if err != nil {
		return err
	}

	s.slavesMu.Lock()
	defer s.slavesMu.Unlock()
	for _, slave := range s.slaves {
		_, err := slave.Write(resp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}
