package main

import (
	"fmt"
)

type ServerInfoSection string

const (
	replication ServerInfoSection = "replication"
)

func (s *ServerV2) replicationInfo() []string {
	info := []string{
		fmt.Sprintf("role:%v", s.role.string()),
	}

	if s.isMaster() {
		info = append(info, fmt.Sprintf("master_replid:%v", s.masterReplId))

		if s.masterReplOffset >= 0 {
			info = append(info, fmt.Sprintf("master_repl_offset:%v", s.masterReplOffset))
		}
	}

	return info
}
