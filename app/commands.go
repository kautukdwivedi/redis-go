package main

import (
	"fmt"
	"net"
	"strings"
)

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
	integer    Type = ':'
)

func (s *server) handleCommand(conn net.Conn, cmd *command) error {
	cmd.parse()

	if s.isMaster() {
		return s.handleCommandOnMaster(conn, cmd)
	} else {
		return s.handleCommandOnSlave(conn, cmd)
	}
}

func (s *server) handleCommandOnMaster(conn net.Conn, cmd *command) error {
	switch strings.ToLower(cmd.name) {
	case "ping":
		return s.handleCommandPing(conn)
	case "echo":
		return s.handleCommandEcho(conn, cmd.args)
	case "get":
		return s.handleCommandGet(conn, cmd.args)
	case "set":
		return s.handleCommandSetOnMaster(conn, cmd.args)
	case "info":
		return s.handleCommandInfo(conn, cmd.args)
	case "replconf":
		return s.handleCommandReplconf(conn)
	case "replconf ack":
		return s.handleCommandReplconfAck()
	case "psync":
		return s.handleCommandPsync(conn)
	case "wait":
		return s.handleCommandWait(conn, cmd.args)
	case "config get":
		return s.handleCommandConfigGet(conn, cmd.args)
	case "keys":
		return s.handleCommandKeys(conn)
	case "incr":
		return s.handleCommandIncr(conn, cmd.args)
	case "multi":
		return s.handleCommandMulti(conn)
	case "exec":
		return s.handleCommandExec(conn)
	default:
		return nil
	}
}

func (s *server) handleCommandOnSlave(conn net.Conn, cmd *command) error {
	var err error

	switch strings.ToLower(cmd.name) {
	case "echo":
		err = s.handleCommandEcho(conn, cmd.args)
	case "get":
		err = s.handleCommandGet(conn, cmd.args)
	case "set":
		err = s.handleCommandSetOnSlave(cmd.args)
	case "info":
		err = s.handleCommandInfo(conn, cmd.args)
	case "replconf getack":
		err = s.handleCommandReplconfGetAck(conn)
	case "config get":
		err = s.handleCommandConfigGet(conn, cmd.args)
	case "keys":
		err = s.handleCommandKeys(conn)
	case "incr":
		err = s.handleCommandIncr(conn, cmd.args)
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
