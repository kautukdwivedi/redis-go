package main

import (
	"errors"
	"flag"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage/rdb"
)

type serverConfig struct {
	port             int
	masterHost       string
	masterPort       int
	masterReplId     string
	masterReplOffset int
	role             ServerRole
	rdbFile          *rdb.RDBFile
}

func newServerConfig() (*serverConfig, error) {
	port := flag.Int("port", 6379, "Server port number")
	replicaOf := flag.String("replicaof", "", "<MASTER_HOST> <MASTER_PORT>")
	rdbFileDir := flag.String("dir", "", "RDB file dir")
	rdbFileName := flag.String("dbfilename", "", "RDB file name")
	flag.Parse()

	config := &serverConfig{
		port:    *port,
		rdbFile: rdb.NewRDBFile(*rdbFileDir, *rdbFileName),
	}

	err := config.parseReplicaOf(*replicaOf)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *serverConfig) parseReplicaOf(replicaOf string) error {
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
