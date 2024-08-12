package main

import (
	"errors"
	"strconv"
	"strings"
)

func (s *server) parseReplicaOf(replicaOf string) error {
	if len(replicaOf) > 0 {
		addrAndPort := strings.Split(replicaOf, " ")

		if len(addrAndPort) != 2 {
			return errors.New("invalid replica address format")
		}

		port, err := strconv.Atoi(addrAndPort[1])
		if err != nil {
			return errors.New("invalid replica port")
		}

		s.options.role = slave
		s.options.masterHost = addrAndPort[0]
		s.options.masterPort = port
		s.options.masterReplId = "?"
		s.options.masterReplOffset = -1
	} else {
		s.options.role = master
		s.options.masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.options.masterReplOffset = 0
	}

	return nil
}
