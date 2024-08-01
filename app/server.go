package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	c, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer c.Close()

	buf := make([]byte, 128)

	for {
		_, err = c.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection: ", err.Error())
			os.Exit(1)
		}

		_, err = c.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing data to connection: ", err.Error())
			os.Exit(1)
		}
	}
}
