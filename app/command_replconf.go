package main

import (
	"strconv"
)

func (s *server) handleCommandReplconf(client *Client) error {
	_, err := client.Write(okSimpleString())
	if err != nil {
		return err
	}
	return nil
}

func (s *server) handleCommandReplconfAck() error {
	s.ackChan <- true
	return nil
}

func (s *server) handleCommandReplconfGetAck(client *Client) error {
	resp, err := respAsArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.masterReplOffset)})
	if err != nil {
		return err
	}

	_, err = client.Write(resp)
	if err != nil {
		return err
	}

	return nil
}
