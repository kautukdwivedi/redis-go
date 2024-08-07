package main

import "time"

type client struct {
	mapData map[string]expiringValue
}

type expiringValue struct {
	val       []byte
	created   time.Time
	expiresIn int
}

func (v *expiringValue) hasExpired() bool {
	return v.expiresIn > 0 && time.Now().UTC().Sub(v.created).Milliseconds() >= int64(v.expiresIn)
}
