package main

import "fmt"

type ServerInfoSection string

const (
	replication ServerInfoSection = "replication"
)

func (s *server) replicationInfo() []string {
	return []string{
		fmt.Sprintf("role:%v", s.role),
	}
}
