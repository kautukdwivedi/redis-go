package main

type ServerRole int

const (
	master = iota + 1
	slave
)

func (r ServerRole) string() string {
	return [...]string{"master", "slave"}[r-1]
}
