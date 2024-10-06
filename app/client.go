package main

import "net"

type Client struct {
	net.Conn
	Transaction *Transaction
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn:        conn,
		Transaction: NewTransaction(conn),
	}
}
