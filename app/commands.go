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

	if client.Transaction.isOpen && cmd.isQueable {
		client.Transaction.Queue = append(client.Transaction.Queue, cmd)
		_, err := client.Write(respAsSimpleString("QUEUED"))
		if err != nil {
			return err
		}
		return nil
	}

	var resp []byte
	var err error

	if s.isMaster() {
		resp, err = s.handleCommandOnMaster(client, cmd)
		if cmd.isWrite {
			go func() {
				err := s.propagateCommandToSlaves(cmd)
				if err != nil {
					fmt.Println("Failed propagating to slaves: ", err)
				}
			}()
		}
	} else {
		resp, err = s.handleCommandOnSlave(client, cmd)
	}
	if err != nil {
		return err
	}

	if len(resp) > 0 {
		_, err = client.Write(resp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleCommandOnMaster(client *Client, cmd *command) (resp []byte, err error) {
	switch cmd.name {
	case "PING":
		return nil, s.handleCommandPing(client)
	case "ECHO":
		return s.handleCommandEcho(cmd.args)
	case "GET":
		return s.handleCommandGet(cmd.args)
	case "SET":
		return s.handleCommandSetOnMaster(cmd.args)
	case "INFO":
		return nil, s.handleCommandInfo(client, cmd.args)
	case "REPLCONF":
		return nil, s.handleCommandReplconf(client)
	case "REPLCONF ACK":
		return nil, s.handleCommandReplconfAck()
	case "PSYNC":
		return nil, s.handleCommandPsync(client)
	case "WAIT":
		return nil, s.handleCommandWait(client, cmd.args)
	case "CONFIG GET":
		return s.handleCommandConfigGet(cmd.args)
	case "KEYS":
		return s.handleCommandKeys()
	case "INCR":
		return s.handleCommandIncr(cmd.args)
	case "MULTI":
		return nil, s.handleCommandMulti(client)
	case "EXEC":
		return nil, s.handleCommandExec(client)
	case "DISCARD":
		return nil, s.handleCommandDiscard(client)
	default:
		return nil, nil
	}
}

func (s *server) handleCommandOnSlave(client *Client, cmd *command) (resp []byte, err error) {
	switch cmd.name {
	case "ECHO":
		resp, err = s.handleCommandEcho(cmd.args)
	case "GET":
		resp, err = s.handleCommandGet(cmd.args)
	case "SET":
		resp, err = s.handleCommandSetOnSlave(cmd.args)
	case "INFO":
		err = s.handleCommandInfo(client, cmd.args)
	case "REPLCONF GETACK":
		err = s.handleCommandReplconfGetAck(client)
	case "INCR":
		resp, err = s.handleCommandIncr(cmd.args)
	}

	if err == nil {
		s.masterReplOffset += cmd.bytesLength()
	}

	return resp, err
}

func (s *server) propagateCommandToSlaves(cmd *command) error {
	argsStr := make([]string, 0, len(cmd.args)+1)
	argsStr = append(argsStr, cmd.name)
	argsStr = append(argsStr, cmd.args...)

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
