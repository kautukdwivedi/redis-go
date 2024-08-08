package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type command struct {
	name     string
	callback func(conn net.Conn, client *client, args [][]byte) error
}

func parseCommand(buf [][]byte) (string, [][]byte) {
	var args [][]byte

	if len(buf) > 2 {
		args = make([][]byte, 0, (len(buf)-2)/2)
		for idx, piece := range buf {
			if idx < 2 {
				continue
			}
			if idx%2 != 0 {
				args = append(args, piece)
			}
		}
	}

	return string(buf[1]), args
}

func (app *server) findCommand(name string) (*command, error) {
	comm, ok := app.getCommands()[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown command: \"%s\"", name)
	}

	return comm, nil
}

func (s *server) getCommands() map[string]*command {
	return map[string]*command{
		"ping": {
			name:     "ping",
			callback: s.handleCommandPing,
		},

		"echo": {
			name:     "echo",
			callback: s.handleCommandEcho,
		},

		"get": {
			name:     "get",
			callback: s.handleCommandGet,
		},

		"set": {
			name:     "set",
			callback: s.handleCommandSet,
		},
		"info": {
			name:     "info",
			callback: s.handleCommandInfo,
		},
	}
}

func (s *server) handleCommandPing(conn net.Conn, client *client, args [][]byte) error {
	_, err := conn.Write(respAsSimpleString("PONG"))
	if err != nil {
		return err
	}

	return nil
}

func (s *server) handleCommandEcho(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	arg := args[0]

	_, err := conn.Write(respAsSimpleString(string(arg)))
	if err != nil {
		return err
	}

	return nil
}

func (s *server) handleCommandGet(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	expVal, ok := client.mapData[string(args[0])]
	if !ok {
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	if expVal.hasExpired() {
		_, err := conn.Write(nullBulkString)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := conn.Write(respAsBulkString(string(expVal.val)))
	if err != nil {
		return err
	}

	return nil
}

func (s *server) handleCommandSet(conn net.Conn, client *client, args [][]byte) error {
	if len(args) < 2 {
		return errors.New("command set accepts two arguments")
	}

	if len(args)%2 != 0 {
		return errors.New("invalid arguments list, must come in pairs")
	}

	key := string(args[0])

	expVal := expiringValue{
		val:     args[1],
		created: time.Now().UTC(),
	}

	if len(args) > 2 {
		extraArg := string(args[2])
		if !strings.EqualFold(extraArg, "px") {
			return fmt.Errorf("unknown extra argument \"%s\"", extraArg)
		}

		exp, err := strconv.Atoi(string(args[3]))
		if err != nil {
			return err
		}

		expVal.expiresIn = exp
	}

	client.mapData[key] = expVal

	_, err := conn.Write(okSimpleString())
	if err != nil {
		return err
	}

	return nil
}

func (s *server) handleCommandInfo(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")

		_, err := conn.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}
