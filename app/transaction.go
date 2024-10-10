package main

import (
	"errors"
	"net"
	"sync"
)

type Transaction struct {
	Conn   net.Conn
	Queue  []*command
	mu     *sync.Mutex
	isOpen bool
}

func NewTransaction(conn net.Conn) *Transaction {
	return &Transaction{
		Conn:  conn,
		Queue: make([]*command, 0),
		mu:    &sync.Mutex{},
	}
}

func (t *Transaction) IsOpen() bool {
	return t.isOpen
}

func (t *Transaction) Open() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isOpen {
		return errors.New("transaction is already open")
	}

	t.isOpen = true

	return nil
}

func (t *Transaction) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isOpen {
		clear(t.Queue)
		t.isOpen = false
		return nil
	}

	return errors.New("transaction is already closed")
}
