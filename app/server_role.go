package main

type severRole int

const (
	master = iota + 1
	slave
)

func (r severRole) string() string {
	return [...]string{"master", "slave"}[r-1]
}
