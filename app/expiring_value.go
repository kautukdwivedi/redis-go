package main

import (
	"fmt"
	"time"
)

type expiringValue struct {
	val       string
	created   time.Time
	expiresIn int
}

func (v *expiringValue) hasExpired() bool {
	fmt.Println("\nExpires in: ", v.expiresIn)
	fmt.Println("Created: ", v.created)
	fmt.Println("Now: ", time.Now().UTC())
	return v.expiresIn > 0 && time.Now().UTC().Sub(v.created).Milliseconds() >= int64(v.expiresIn)
}
