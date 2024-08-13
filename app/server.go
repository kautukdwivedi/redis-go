package main

import (
	"fmt"
	"net"
)

type Type byte

const (
	array      Type = '*'
	bulkString Type = '$'
)

type server struct {
	options *serverOptions
}

type serverOptions struct {
	role             severRole
	port             int
	masterHost       string
	masterPort       int
	masterReplId     string
	masterReplOffset int
	replicas         []*net.Conn
}

func (s *server) isMaster() bool {
	return s.options.role == master
}

func (s *server) isSlave() bool {
	return s.options.role == slave
}

func (s *server) propagateCommandToReplicas(comm string, args [][]byte) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)

	for _, arg := range args {
		argsStr = append(argsStr, string(arg))
	}

	resp, err := respAsArray(argsStr)
	if err != nil {
		return nil
	}

	for _, r := range s.options.replicas {
		_, err := (*r).Write(resp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}
