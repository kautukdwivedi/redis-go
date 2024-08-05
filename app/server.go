package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
)

type application struct {
	commands map[string]*command
}

type client struct {
	mapData map[string][]byte
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	app := &application{
		commands: getCommands(),
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		client := &client{
			mapData: map[string][]byte{},
		}

		go app.handleClient(conn, client)
	}
}

func (app *application) handleClient(conn net.Conn, client *client) {
	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			return
		}

		fmt.Println("Values read: ", strings.Split(string(buf), "\r\n"))

		err = app.parseInputBuffer(buf, conn, client)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (app *application) parseInputBuffer(buf []byte, conn net.Conn, client *client) error {
	if len(buf) == 0 {
		return errors.New("empty input buffer")
	}

	if len(buf) == 1 {
		return errors.New("input only contains data type information, but no data")
	}

	var t Type = Type(buf[0])
	switch t {
	case array:
		splitBuf := bytes.Split(buf, []byte("\r\n"))

		if len(splitBuf) == 1 {
			return errors.New("input data is an array, but does not contain actual data")
		}

		name, args := parseCommand(splitBuf[1:])

		comm, err := app.findCommand(name)
		if err != nil {
			return err
		}

		comm.callback(conn, client, args)
	}
	return nil
}
