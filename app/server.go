package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	port             int
	data             = make(map[string]expiringValue)
	dataMu           sync.RWMutex
	slaves           []net.Conn
	slavesMu         sync.Mutex
	masterHost       string
	masterPort       int
	masterReplId     string
	masterReplOffset int
	role             ServerRole
)

func startServer() {
	if role == slave {
		masterConn, err := doHandshakeWithMaster()
		if err != nil {
			fmt.Println("Error connecting to master: ", err)
			return
		}

		go handleConn(masterConn)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading from conn: ", err)
			continue
		}

		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		handleRawMessage(conn, msgBuf)
	}
}

func serverConfig() error {
	p := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "<MASTER_HOST> <MASTER_PORT>")
	flag.Parse()

	port = *p

	err := parseReplicaOf(*replicaOf)
	if err != nil {
		return err
	}

	return nil
}

func parseReplicaOf(replicaOf string) error {
	if len(replicaOf) > 0 {
		addrAndPort := strings.Split(replicaOf, " ")

		if len(addrAndPort) != 2 {
			return errors.New("invalid replica address format")
		}

		port, err := strconv.Atoi(addrAndPort[1])
		if err != nil {
			return errors.New("invalid replica port")
		}

		masterHost = addrAndPort[0]
		masterPort = port
		masterReplId = "?"
		masterReplOffset = -1

		role = slave
	} else {
		masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		masterReplOffset = 0

		role = master
	}

	return nil
}

func handleRawMessage(conn net.Conn, msgBuf []byte) error {
	bufStr := string(msgBuf)
	if len(bufStr) == 0 {
		return errors.New("empty raw message")
	}

	commands := strings.Split(bufStr, "*")
	if len(commands) == 0 {
		return errors.New("no commands in input message")
	}

	for _, command := range commands[1:] {
		err := handleCommand(conn, command)
		if err != nil {
			fmt.Println("cmd error: ", err)
		}
	}

	return nil
}

func handleCommand(conn net.Conn, cmd string) error {
	cmdPieces := strings.Split(cmd, carriageReturn())

	if len(cmdPieces) <= 1 {
		return errors.New("command string is not a valid command")
	}

	name, args := parseCommand(cmdPieces[1:])
	comm, err := findCommand(name)

	if err != nil {
		return err
	}

	comm.callback(conn, args)

	return nil
}

func propagateCommandToSlaves(comm string, args []string) error {
	argsStr := make([]string, 0, len(args)+1)
	argsStr = append(argsStr, comm)
	argsStr = append(argsStr, args...)

	resp, err := respAsArray(argsStr)
	if err != nil {
		return err
	}

	fmt.Printf("Propagating commands to %d slaves\n", len(slaves))

	slavesMu.Lock()
	defer slavesMu.Unlock()
	for _, slave := range slaves {
		_, err := slave.Write(resp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}
