package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type command struct {
	name     string
	callback func(client *ClientV2, args []string) error
}

func parseCommand(cmdPieces []string) (string, []string) {
	cleanCmdPieces := make([]string, 0, len(cmdPieces)/2-1)
	for idx, piece := range cmdPieces {
		if idx%2 != 0 {
			cleanCmdPieces = append(cleanCmdPieces, piece)
		}
	}

	fmt.Println("Parsing clean command: ", strings.Join(cleanCmdPieces, ","))

	var args []string

	if len(cleanCmdPieces) > 1 {
		args = cleanCmdPieces[1:]
	}

	return string(cleanCmdPieces[0]), args
}

func (app *ServerV2) findCommand(name string) (*command, error) {
	comm, ok := app.getCommands()[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown command: \"%s\"", name)
	}

	return comm, nil
}

func (s *ServerV2) getCommands() map[string]*command {
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
		"replconf": {
			name:     "replconf",
			callback: s.handleCommandReplconf,
		},
		"psync": {
			name:     "psync",
			callback: s.handleCommandPsync,
		},
	}
}

func (s *ServerV2) handleCommandPing(client *ClientV2, args []string) error {
	_, err := client.conn.Write(respAsSimpleString("PONG"))
	if err != nil {
		fmt.Println("Received error in PING command: ", err)
		return err
	}

	return nil
}

func (s *ServerV2) handleCommandEcho(client *ClientV2, args []string) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	_, err := client.conn.Write(respAsSimpleString(args[0]))
	if err != nil {
		fmt.Println("Received error in ECHO command: ", err)
		return err
	}

	return nil
}

func (s *ServerV2) handleCommandGet(client *ClientV2, args []string) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	fmt.Println("Getting val for key: ", args[0])

	expVal, ok := client.Get(args[0])
	if !ok {
		_, err := client.conn.Write(nullBulkString)
		if err != nil {
			fmt.Println("Received error in GET command: ", err)
			return err
		}

		return nil
	}

	if expVal.hasExpired() {
		_, err := client.conn.Write(nullBulkString)
		if err != nil {
			fmt.Println("Received error in GET1 command: ", err)
			return err
		}

		return nil
	}

	_, err := client.conn.Write(respAsBulkString(string(expVal.val)))
	if err != nil {
		fmt.Println("Received error in GET2 command: ", err)
		return err
	}

	return nil
}

func (s *ServerV2) handleCommandSet(client *ClientV2, args []string) error {
	fmt.Println("Handlign command set with args: ", args)

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
		extraArg := args[2]
		if !strings.EqualFold(extraArg, "px") {
			return fmt.Errorf("unknown extra argument \"%s\"", extraArg)
		}

		exp, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println("Received error in SET command: ", err)
			return err
		}

		expVal.expiresIn = exp
	}

	client.Set(context.Background(), key, expVal)

	sleepSeconds(1)

	if s.isMaster() {
		_, err := client.conn.Write(okSimpleString())
		if err != nil {
			fmt.Println("Received error in SET1 command: ", err)
			return err
		}

		err = s.propagateCommandToReplicas("SET", args)
		if err != nil {
			fmt.Println("Received error in SET2 command: ", err)
			return err
		}
	}

	return nil
}

func (s *ServerV2) handleCommandInfo(client *ClientV2, args []string) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")

		_, err := client.conn.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerV2) handleCommandReplconf(client *ClientV2, args []string) error {
	_, err := client.conn.Write(okSimpleString())
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerV2) handleCommandPsync(client *ClientV2, args []string) error {
	resp := fmt.Sprintf("FULLRESYNC %s %d", s.masterReplId, s.masterReplOffset)

	_, err := client.conn.Write(respAsSimpleString(resp))
	if err != nil {
		return err
	}

	sleepSeconds(1)

	emptyFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	fileData, err := base64.StdEncoding.DecodeString(emptyFileBase64)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(respAsFileData(fileData))
	if err != nil {
		return err
	}

	fmt.Println("Adding replica...")
	s.replicas = append(s.replicas, &client.conn)

	return nil
}
