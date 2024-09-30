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
	return v.ExpiresIn > 0 && time.Now().UTC().Sub(v.Created).Milliseconds() >= int64(v.ExpiresIn)
}
