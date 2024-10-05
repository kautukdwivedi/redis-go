package storage

import (
	"time"
)

type ExpiringValue struct {
	Val       string
	Created   time.Time
	ExpiresIn int
}

func (v *ExpiringValue) HasExpired() bool {
	if v.ExpiresIn < 0 {
		return true
	}
	if v.ExpiresIn == 0 {
		return false
	}
	return time.Now().UTC().Sub(v.Created).Milliseconds() >= int64(v.ExpiresIn)
}
