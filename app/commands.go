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

const (
	cmdPing           = "ping"
	cmdEcho           = "echo"
	cmdGet            = "get"
	cmdSet            = "set"
	cmdInfo           = "info"
	cmdReplConf       = "replconf"
	cmdReplConfGetAck = "replconf getack"
	cmdReplConfAck    = "replconf ack"
	cmdPsync          = "psync"
	cmdWait           = "wait"
	cmdConfigGet      = "config get"
	cmdKeys           = "keys"
	cmdIncr           = "incr"
	cmdMulti          = "multi"
	cmdExec           = "exec"
	cmdDiscard        = "discard"
	cmdType           = "type"
	cmdXAdd           = "xadd"
	cmdXRange         = "xrange"
	cmdXRead          = "xread"
)

var supportedCommands = []string{
	cmdPing,
	cmdEcho,
	cmdGet,
	cmdSet,
	cmdInfo,
	cmdReplConf,
	cmdReplConfGetAck,
	cmdReplConfAck,
	cmdPsync,
	cmdWait,
	cmdConfigGet,
	cmdKeys,
	cmdIncr,
	cmdMulti,
	cmdExec,
	cmdDiscard,
	cmdType,
	cmdXAdd,
	cmdXRange,
	cmdXRead,
}

func (s *server) handleCommand(client *Client, cmd *command) error {
	err := cmd.parse()
	if err != nil {
		return err
	}

	if client.Transaction.isOpen && cmd.isQueable {
		client.Transaction.Queue = append(client.Transaction.Queue, cmd)
		_, err := client.Write(respAsSimpleString("QUEUED"))
		if err != nil {
			return err
		}
		return nil
	}

	var resp []byte

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
	case cmdPing:
		return nil, s.handleCommandPing(client)
	case cmdEcho:
		return s.handleCommandEcho(cmd.args)
	case cmdGet:
		return s.handleCommandGet(cmd.args)
	case cmdSet:
		return s.handleCommandSetOnMaster(cmd.args)
	case cmdInfo:
		return nil, s.handleCommandInfo(client, cmd.args)
	case cmdReplConf:
		return nil, s.handleCommandReplconf(client)
	case cmdReplConfAck:
		return nil, s.handleCommandReplconfAck()
	case cmdPsync:
		return nil, s.handleCommandPsync(client)
	case cmdWait:
		return nil, s.handleCommandWait(client, cmd.args)
	case cmdConfigGet:
		return s.handleCommandConfigGet(cmd.args)
	case cmdKeys:
		return s.handleCommandKeys()
	case cmdIncr:
		return s.handleCommandIncr(cmd.args)
	case cmdMulti:
		return nil, s.handleCommandMulti(client)
	case cmdExec:
		return nil, s.handleCommandExec(client)
	case cmdDiscard:
		return nil, s.handleCommandDiscard(client)
	case cmdType:
		return s.handleCommandType(cmd.args)
	case cmdXAdd:
		return s.handleCommandXADD(cmd.args)
	case cmdXRange:
		return s.handleCommandXRANGE(cmd.args)
	case cmdXRead:
		return s.handleCommandXREAD(cmd.args)
	default:
		return nil, nil
	}
}

func (s *server) handleCommandOnSlave(client *Client, cmd *command) (resp []byte, err error) {
	switch cmd.name {
	case cmdEcho:
		resp, err = s.handleCommandEcho(cmd.args)
	case cmdGet:
		resp, err = s.handleCommandGet(cmd.args)
	case cmdSet:
		resp, err = s.handleCommandSetOnSlave(cmd.args)
	case cmdInfo:
		err = s.handleCommandInfo(client, cmd.args)
	case cmdReplConfGetAck:
		err = s.handleCommandReplconfGetAck(client)
	case cmdIncr:
		resp, err = s.handleCommandIncr(cmd.args)
	case cmdType:
		resp, err = s.handleCommandType(cmd.args)
	case cmdXRange:
		resp, err = s.handleCommandXRANGE(cmd.args)
	case cmdXRead:
		resp, err = s.handleCommandXREAD(cmd.args)
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
