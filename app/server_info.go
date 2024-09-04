package main

import (
	"fmt"
)

type ServerInfoSection string

const (
	replication ServerInfoSection = "replication"
)

func replicationInfo() []string {
	info := []string{
		fmt.Sprintf("role:%v", role.string()),
	}

	if role == master {
		info = append(info, fmt.Sprintf("master_replid:%v", masterReplId))

		if masterReplOffset >= 0 {
			info = append(info, fmt.Sprintf("master_repl_offset:%v", masterReplOffset))
		}
	}

	return info
}
