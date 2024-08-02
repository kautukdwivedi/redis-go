package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleClient(c)
	}
}

func handleClient(conn net.Conn) {
	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			return
		}

		fmt.Println("Values read: ", strings.Split(string(buf), "\r\n"))

		parseInputBuffer(buf, conn)
	}
}

func parseInputBuffer(buf []byte, conn net.Conn) error {
	if len(buf) == 0 {
		return errors.New("empty input buffer")
	}

	if len(buf) == 1 {
		return errors.New("input only contains data type information, but no data")
	}

	switch buf[0] {
	case '*':
		splitBuf := bytes.Split(buf, []byte("\r\n"))

		if len(splitBuf) == 1 {
			return errors.New("input data is an array, but does not contain actual data")
		}

		comm := newCommand(splitBuf[1:])
		fmt.Println("Command: ", comm)

		handleCommamd(comm, conn)
	}
	return nil
}

type command struct {
	Command string
	Args    []any
}

func newCommand(buf [][]byte) *command {
	comm := &command{
		Command: string(buf[1]),
	}

	if len(buf) > 2 {
		args := make([]any, 0, (len(buf)-2)/2)
		for idx, piece := range buf {
			if idx < 2 {
				continue
			}
			if idx%2 != 0 {
				args = append(args, string(piece))
			}
		}
		comm.Args = args
	}

	return comm
}

func handleCommamd(c *command, conn net.Conn) error {
	switch strings.ToLower(c.Command) {
	case "echo":
		if len(c.Args) == 1 {
			_, err := conn.Write([]byte(fmt.Sprintf("+%s\r\n", c.Args...)))
			if err != nil {
				fmt.Println("Error writing data to connection: ", err.Error())
			}
			break
		}

		return errors.New("command echo only accepts one argument")

	case "ping":
		_, err := conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing data to connection: ", err.Error())
		}
	}

	return nil
}
