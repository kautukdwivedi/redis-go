package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
)

type command struct {
	name     string
	callback func(conn net.Conn, args []string) error
}

func parseCommand(cmdPieces []string) (string, []string) {
	cleanCmdPieces := make([]string, 0, len(cmdPieces)/2-1)
	for idx, piece := range cmdPieces {
		if idx%2 != 0 {
			cleanCmdPieces = append(cleanCmdPieces, piece)
		}
	}

	var args []string

	if len(cleanCmdPieces) > 1 {
		args = cleanCmdPieces[1:]
	}

	return string(cleanCmdPieces[0]), args
}

func findCommand(name string) (*command, error) {
	comm, ok := getCommands()[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown command: \"%s\"", name)
	}

	return comm, nil
}

func getCommands() map[string]*command {
	return map[string]*command{
		"ping": {
			name:     "ping",
			callback: handleCommandPing,
		},

		"echo": {
			name:     "echo",
			callback: handleCommandEcho,
		},

		"get": {
			name:     "get",
			callback: handleCommandGet,
		},

		"set": {
			name:     "set",
			callback: handleCommandSet,
		},
		"info": {
			name:     "info",
			callback: handleCommandInfo,
		},
		"replconf": {
			name:     "replconf",
			callback: handleCommandReplconf,
		},
		"psync": {
			name:     "psync",
			callback: handleCommandPsync,
		},
	}
}

func handleCommandPing(conn net.Conn, args []string) error {
	_, err := conn.Write(respAsSimpleString("PONG"))
	if err != nil {
		return err
	}

	return nil
}

func handleCommandEcho(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	_, err := conn.Write(respAsSimpleString(args[0]))
	if err != nil {
		return err
	}

	return nil
}

func handleCommandGet(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	dataMu.RLock()
	expVal, ok := data[args[0]]
	dataMu.RUnlock()
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

func handleCommandSet(conn net.Conn, args []string) error {
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
			return err
		}

		expVal.expiresIn = exp
	}

	dataMu.Lock()
	data[key] = expVal
	dataMu.Unlock()

	if role == master {
		_, err := conn.Write(okSimpleString())
		if err != nil {
			return err
		}

		go func() {
			err := propagateCommandToSlaves("SET", args)
			if err != nil {
				fmt.Println("Failed propagating to slaves: ", err)
			}
		}()
	}

	return nil
}

func handleCommandInfo(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(replicationInfo(), "\n")

		_, err := conn.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}

func handleCommandReplconf(conn net.Conn, args []string) error {
	_, err := conn.Write(okSimpleString())
	if err != nil {
		return err
	}

	return nil
}

func handleCommandPsync(conn net.Conn, args []string) error {
	resp := fmt.Sprintf("FULLRESYNC %s %d", masterReplId, masterReplOffset)

	_, err := conn.Write(respAsSimpleString(resp))
	if err != nil {
		return err
	}

	emptyFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	fileData, err := base64.StdEncoding.DecodeString(emptyFileBase64)
	if err != nil {
		return err
	}

	_, err = conn.Write(respAsFileData(fileData))
	if err != nil {
		return err
	}

	fmt.Println("Adding slave...")
	slavesMu.Lock()
	slaves = append(slaves, conn)
	slavesMu.Unlock()

	dataMu.RLock()
	for k, v := range data {
		resp := make([]string, 0, 5)
		resp = append(resp, "SET")
		resp = append(resp, k)
		resp = append(resp, v.val)
		if px := v.expiresIn; px > 0 {
			resp = append(resp, "px")
			resp = append(resp, string(intToByteSlice(px)))
		}
		r, err := respAsArray(resp)
		if err != nil {
			continue
		}

		_, err = conn.Write(r)
		if err != nil {
			return err
		}
	}
	dataMu.RUnlock()

	return nil
}
