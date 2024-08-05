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

		data := &appData{
			mapData: map[string]any{},
		}

		go handleClient(c, data)
	}
}

func handleClient(conn net.Conn, data *appData) {
	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			return
		}

		fmt.Println("Values read: ", strings.Split(string(buf), "\r\n"))

		parseInputBuffer(buf, conn, data)
	}
}

func parseInputBuffer(buf []byte, conn net.Conn, data *appData) error {
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

		handleCommamd(comm, conn, data)
	}
	return nil
}

type command struct {
	Command string
	Args    []any
}

type appData struct {
	mapData map[string]any
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

func handleCommamd(c *command, conn net.Conn, data *appData) error {
	switch strings.ToLower(c.Command) {
	case "echo":
		if len(c.Args) == 1 {
			_, err := conn.Write([]byte(fmt.Sprintf("+%s\r\n", c.Args...)))
			if err != nil {
				fmt.Println("Error writing data to connection: ", err.Error())
			}

			break
		}

		return errors.New("command echo must take one argument")

	case "ping":
		_, err := conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing data to connection: ", err.Error())
		}

	case "get":
		if len(c.Args) == 1 {
			key, ok := c.Args[0].(string)
			if !ok {
				return errors.New("key arg is not a string")
			}

			val, ok := data.mapData[key]
			if !ok {
				_, err := conn.Write([]byte("$-1\r\n"))
				if err != nil {
					fmt.Println("Error writing data to connection: ", err.Error())
				}

				break
			}

			strVal, ok := val.(string)
			if !ok {
				_, err := conn.Write([]byte("$-1\r\n"))
				if err != nil {
					fmt.Println("Error writing data to connection: ", err.Error())
				}

				break
			}

			encoded := fmt.Sprintf("$%d\r\n%v\r\n", len(strVal), strVal)
			_, err := conn.Write([]byte(encoded))
			if err != nil {
				fmt.Println("Error writing data to connection: ", err.Error())
			}

			break
		}

		return errors.New("command get must take one argument")

	case "set":
		if len(c.Args) == 2 {
			key, ok := c.Args[0].(string)
			if !ok {
				return errors.New("key arg is not a string")
			}
			data.mapData[key] = c.Args[1]
			_, err := conn.Write([]byte("+OK\r\n"))
			if err != nil {
				fmt.Println("Error writing data to connection: ", err.Error())
			}

			break
		}

		return errors.New("command set accepts two arguments")
	}

	return nil
}
