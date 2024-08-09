package main

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
	masterHost       string
	masterPort       int
	masterReplId     string
	masterReplOffset int
}

func (s *server) isMaster() bool {
	return s.options.role == master
}

func (s *server) isSlave() bool {
	return s.options.role == slave
}
