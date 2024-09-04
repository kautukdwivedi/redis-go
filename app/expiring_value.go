package main

import (
	"time"
)

type expiringValue struct {
	val       string
	created   time.Time
	expiresIn int
}

func (v *expiringValue) hasExpired() bool {
	return v.expiresIn > 0 && time.Now().UTC().Sub(v.created).Milliseconds() >= int64(v.expiresIn)
}
