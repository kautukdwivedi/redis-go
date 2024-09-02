package main

import (
	"errors"
	"strconv"
	"strings"
)

func (s *server) parseReplicaOf(replicaOf string) (*ServerRole, error) {
	var role ServerRole
	if len(replicaOf) > 0 {
		addrAndPort := strings.Split(replicaOf, " ")

		if len(addrAndPort) != 2 {
			return nil, errors.New("invalid replica address format")
		}

		port, err := strconv.Atoi(addrAndPort[1])
		if err != nil {
			return nil, errors.New("invalid replica port")
		}

		s.options.masterHost = addrAndPort[0]
		s.options.masterPort = port
		s.options.masterReplId = "?"
		s.options.masterReplOffset = -1

		role = slave
	} else {
		s.options.masterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.options.masterReplOffset = 0

		role = master
	}
	s.options.role = master

	return &role, nil
}
