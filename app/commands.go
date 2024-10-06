package main

import (
	"fmt"
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
		if client.Transaction.isOpen {
			if cmd.name != "MULTI" && cmd.name != "EXEC" {
				client.Transaction.Queue = append(client.Transaction.Queue, cmd)
				_, err := client.Write(respAsSimpleString("QUEUED"))
				if err != nil {
					return err
				}
				return nil
			}
		}
		return s.handleCommandOnMaster(client, cmd)
	} else {
		return s.handleCommandOnSlave(client, cmd)
	}
}

func (s *server) handleCommandOnMaster(client *Client, cmd *command) error {
	switch cmd.name {
	case "PING":
		return s.handleCommandPing(client)
	case "ECHO":
		return s.handleCommandEcho(client, cmd.args)
	case "GET":
		return s.handleCommandGet(client, cmd.args)
	case "SET":
		return s.handleCommandSetOnMaster(client, cmd.args)
	case "INFO":
		return s.handleCommandInfo(client, cmd.args)
	case "REPLCONF":
		return s.handleCommandReplconf(client)
	case "REPLCONF ACK":
		return s.handleCommandReplconfAck()
	case "PSYNC":
		return s.handleCommandPsync(client)
	case "WAIT":
		return s.handleCommandWait(client, cmd.args)
	case "CONFIG GET":
		return s.handleCommandConfigGet(client, cmd.args)
	case "KEYS":
		return s.handleCommandKeys(client)
	case "INCR":
		return s.handleCommandIncr(client, cmd.args)
	case "MULTI":
		return s.handleCommandMulti(client)
	case "EXEC":
		return s.handleCommandExec(client)
	default:
		return nil
	}
}

func (s *server) handleCommandOnSlave(client *Client, cmd *command) error {
	var err error

	switch cmd.name {
	case "ECHO":
		err = s.handleCommandEcho(client, cmd.args)
	case "GET":
		err = s.handleCommandGet(client, cmd.args)
	case "SET":
		err = s.handleCommandSetOnSlave(cmd.args)
	case "INFO":
		err = s.handleCommandInfo(client, cmd.args)
	case "REPLCONF GETACK":
		err = s.handleCommandReplconfGetAck(client)
	case "CONFIG GET":
		err = s.handleCommandConfigGet(client, cmd.args)
	case "KEYS":
		err = s.handleCommandKeys(client)
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
