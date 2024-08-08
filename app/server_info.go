package main

import "fmt"

type ServerInfoSection string

const (
	replication ServerInfoSection = "replication"
)

func replicationInfo() []string {
	return []string{
		fmt.Sprintf("role:%v", role()),
	}
}

func role() string {
	return "master"
}
