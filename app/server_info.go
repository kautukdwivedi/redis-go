package main

import (
	"fmt"
)

type ServerInfoSection string

const (
	replication ServerInfoSection = "replication"
)

func (s *server) replicationInfo() []string {
	info := []string{
		fmt.Sprintf("role:%v", s.options.role.string()),
	}

	if s.isMaster() {
		info = append(info, fmt.Sprintf("master_replid:%v", s.options.masterReplId))

		if s.options.masterReplOffset >= 0 {
			info = append(info, fmt.Sprintf("master_repl_offset:%v", s.options.masterReplOffset))
		}
	}

	return info
}
