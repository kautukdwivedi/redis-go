package main

import (
	"strconv"
)

func (s *server) handleCommandReplconf(client *Client) ([]byte, error) {
	_, err := client.Write(okSimpleString())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *server) handleCommandReplconfAck() ([]byte, error) {
	s.ackChan <- true
	return nil, nil
}

func (s *server) handleCommandReplconfGetAck(client *Client) ([]byte, error) {
	resp, err := respAsArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.masterReplOffset)})
	if err != nil {
		return nil, err
	}

	_, err = client.Write(resp)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
