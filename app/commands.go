package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
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

func (app *application) findCommand(name string) (*command, error) {
	comm, ok := app.commands[strings.ToLower(name)]
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
	}
}

func handleCommandPing(conn net.Conn, client *client, args [][]byte) error {
	_, err := conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func handleCommandEcho(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	arg := args[0]

	_, err := conn.Write([]byte(fmt.Sprintf("+%s\r\n", string(arg))))
	if err != nil {
		return err
	}

	return nil
}

func handleCommandGet(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	val, ok := client.mapData[string(args[0])]
	if !ok {
		_, err := conn.Write([]byte("$-1\r\n"))

		if err != nil {
			return err
		}
	}

	encoded := fmt.Sprintf("$%d\r\n%v\r\n", len(val), string(val))

	_, err := conn.Write([]byte(encoded))
	if err != nil {
		return err
	}

	return nil
}

func handleCommandSet(conn net.Conn, client *client, args [][]byte) error {
	if len(args) != 2 {
		return errors.New("command set accepts two arguments")
	}

	key := string(args[0])

	client.mapData[key] = args[1]

	_, err := conn.Write([]byte("+OK\r\n"))
	if err != nil {
		return err
	}

	return nil
}
