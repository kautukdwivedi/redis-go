package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

type ConfigV2 struct {
	port             int
	role             ServerRole
	masterHost       string
	masterPort       int
	masterReplId     string
	masterReplOffset int
}

type ServerV2 struct {
	*ConfigV2
	peers     []*ClientV2
	ln        net.Listener
	addPeerCh chan *ClientV2
	quitCh    chan struct{}
	msgCh     chan Message
	replicas  []*net.Conn
}

func (s *ServerV2) isMaster() bool {
	return s.role == master
}

func (s *ServerV2) isSlave() bool {
	return s.role == slave
}

func NewServer(cfg *ConfigV2) *ServerV2 {
	return &ServerV2{
		ConfigV2:  cfg,
		peers:     []*ClientV2{},
		addPeerCh: make(chan *ClientV2),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		replicas:  []*net.Conn{},
	}
}

func (s *ServerV2) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	if s.isSlave() {
		conn, err := s.doHandshakeWithMaster()
		if err != nil {
			log.Fatal("Could not handshake with master")
		}

		go s.handleConn(conn)
		return nil
	}

	go s.acceptLoop()

	return nil
}

func (s *ServerV2) loop() {
	for {
		select {
		case peer := <-s.addPeerCh:
			s.peers = append(s.peers, peer)
		case <-s.quitCh:
			return
		case rawMsg := <-s.msgCh:
			s.handleRawMessage(rawMsg)
		}
	}
}

func (s *ServerV2) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *ServerV2) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err)
	}
}

func serverConfig() (*ConfigV2, error) {
	port := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "<MASTER_HOST> <MASTER_PORT>")
	flag.Parse()

	config := &ConfigV2{}

	config.port = *port

	err := config.parseReplicaOf(*replicaOf)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *ConfigV2) parseReplicaOf(replicaOf string) error {
	if len(replicaOf) > 0 {
		addrAndPort := strings.Split(replicaOf, " ")

		if len(addrAndPort) != 2 {
			return errors.New("invalid replica address format")
		}

		port, err := strconv.Atoi(addrAndPort[1])
		if err != nil {
			return errors.New("invalid replica port")
		}

		c.masterHost = addrAndPort[0]
		c.masterPort = port
		c.masterReplId = "?"
		c.masterReplOffset = -1

		c.role = slave
	} else {
		c.masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		c.masterReplOffset = 0

		c.role = master
	}

	return nil
}

func (s *ServerV2) handleRawMessage(msg Message) error {
	bufStr := string(msg.msgBuf)
	if len(bufStr) == 0 {
		return errors.New("empty raw message")
	}

	commands := strings.Split(bufStr, "*")
	if len(commands) == 0 {
		return errors.New("no commands in input message")
	}

	for _, command := range commands[1:] {
		err := s.handleCommand(msg.client, command)
		if err != nil {
			fmt.Println("cmd error: ", err)
		}
	}

	return nil

	// buf := msg.msgBuf
	// if len(buf) == 0 {
	// 	return errors.New("empty input buffer")
	// }

	// if len(buf) == 1 {
	// 	return errors.New("input only contains data type information, but no data")
	// }

	// fmt.Println("Reading bytes: ", string(buf))

	// var t Type = Type(buf[0])
	// switch t {
	// case array:
	// 	splitBuf := bytes.Split(buf, []byte("\r\n"))

	// 	if len(splitBuf) == 1 {
	// 		return errors.New("input data is an array, but does not contain actual data")
	// 	}

	// 	name, args := parseCommand(splitBuf[1:])

	// 	comm, err := s.findCommand(name)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	comm.callback(msg.client, args)
	// }
	// return nil
}

func (s *ServerV2) handleCommand(client *ClientV2, cmd string) error {
	cmdPieces := strings.Split(cmd, carriageReturn())

	fmt.Println("Parsing command: ", cmdPieces[1:])

	if len(cmdPieces) <= 1 {
		return errors.New("command string is not a valid command")
	}

	name, args := parseCommand(cmdPieces[1:])
	comm, err := s.findCommand(name)

	if err != nil {
		return err
	}

	comm.callback(client, args)

	return nil
}

func (s *ServerV2) propagateCommandToReplicas(comm string, args []string) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)
	argsStr = append(argsStr, args...)

	resp, err := respAsArray(argsStr)
	if err != nil {
		return err
	}

	fmt.Printf("Propagating commands to %d replicas", len(s.replicas))

	for _, r := range s.replicas {
		_, err := (*r).Write(resp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}
