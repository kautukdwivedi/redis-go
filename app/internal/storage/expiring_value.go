package storage

import (
	"time"
)

type ExpiringValue struct {
	Val       string
	Created   time.Time
	ExpiresIn int
}

func NewExpiringValue(val string) *ExpiringValue {
	return &ExpiringValue{
		Val:     val,
		Created: time.Now().UTC(),
	}
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
