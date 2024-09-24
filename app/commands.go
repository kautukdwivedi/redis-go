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

func (s *server) handleCommandPing(conn net.Conn) error {
	if s.isMaster() {
		_, err := conn.Write(respAsSimpleString("PONG"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleCommandEcho(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command echo must take one argument")
	}

	_, err := conn.Write(respAsSimpleString(args[0]))
	if err != nil {
		return err
	}

	return nil
}

func (s *server) handleCommandGet(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("command get must take one argument")
	}

	nullBulkString := respAsBulkString("")

	s.dataMu.RLock()
	expVal, ok := s.data[args[0]]
	s.dataMu.RUnlock()
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

func (s *server) handleCommandSet(conn net.Conn, args []string) error {
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

	s.dataMu.Lock()
	s.data[key] = expVal
	s.dataMu.Unlock()

	if s.isMaster() {
		_, err := conn.Write(okSimpleString())
		if err != nil {
			return err
		}

		go func() {
			err := s.propagateCommandToSlaves("SET", args)
			if err != nil {
				fmt.Println("Failed propagating to slaves: ", err)
			}
		}()
	}

	return nil
}

func (s *server) handleCommandInfo(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return errors.New("not yet supported")
	}

	switch ServerInfoSection(args[0]) {
	case replication:
		respStr := strings.Join(s.replicationInfo(), "\n")

		_, err := conn.Write(respAsBulkString(respStr))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleCommandReplconf(conn net.Conn, args []string) error {
	if len(args) == 2 && strings.ToLower(args[0]) == "getack" && args[1] == "*" {
		resp, err := respAsArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.masterReplOffset)})
		if err != nil {
			return err
		}

		_, err = conn.Write(resp)
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Write(okSimpleString())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleCommandPsync(conn net.Conn) error {
	resp := fmt.Sprintf("FULLRESYNC %s %d", s.masterReplId, s.masterReplOffset)

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
	s.slavesMu.Lock()
	s.slaves = append(s.slaves, conn)
	s.slavesMu.Unlock()

	s.dataMu.RLock()
	for k, v := range s.data {
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
	s.dataMu.RUnlock()

	return nil
}
