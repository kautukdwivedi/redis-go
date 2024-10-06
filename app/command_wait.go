package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func (s *server) handleCommandWait(client *Client, args []string) error {
	if len(s.data) == 0 {
		_, err := client.Write(respAsInteger(len(s.slaves)))
		if err != nil {
			return err
		}

	} else {
		for _, slave := range s.slaves {
			go getAckFromSlave(slave)
		}

		requestedSlaves, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		timeout, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		acks := 0
		timer := time.After(time.Duration(timeout) * time.Millisecond)

	outer:
		for acks < requestedSlaves {
			select {
			case <-s.ackChan:
				acks++
			case <-timer:
				break outer
			}
		}

		_, err = client.Write(respAsInteger(acks))
		if err != nil {
			return err
		}
	}

	return nil
}

func getAckFromSlave(slave net.Conn) {
	getAck, err := respAsArray([]string{"REPLCONF", "GETACK", "*"})
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = slave.Write(getAck)
	if err != nil {
		fmt.Println(err.Error())
	}
}
