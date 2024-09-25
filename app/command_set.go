package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func (s *server) handleCommandSetOnMaster(conn net.Conn, args []string) error {
	err := s.handleCommandSet(args)
	if err != nil {
		return err
	}

	_, err = conn.Write(okSimpleString())
	if err != nil {
		return err
	}

	go func() {
		err := s.propagateCommandToSlaves("SET", args)
		if err != nil {
			fmt.Println("Failed propagating to slaves: ", err)
		}
	}()
	return nil
}

func (s *server) handleCommandSetOnSlave(args []string) error {
	return s.handleCommandSet(args)
}

func (s *server) handleCommandSet(args []string) error {
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

	return nil
}
