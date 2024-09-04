package main

import (
	"context"
	"fmt"

	"net"
	"sync"
)

type Message struct {
	client *ClientV2
	msgBuf []byte
}

type ClientV2 struct {
	conn  net.Conn
	msgCh chan *Message
	data  map[string]expiringValue
	mu    *sync.RWMutex
}

func NewPeer(conn net.Conn, msgCh chan *Message) *ClientV2 {
	return &ClientV2{
		conn:  conn,
		msgCh: msgCh,
		data:  make(map[string]expiringValue),
		mu:    &sync.RWMutex{},
	}
}

func (c *ClientV2) readLoop() error {
	buf := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			return err
		}

		fmt.Println("Received new data: ")

		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		c.msgCh <- &Message{
			msgBuf: msgBuf,
			client: c,
		}
	}
}

func (c *ClientV2) Set(ctx context.Context, key string, val expiringValue) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = val
}

func (c *ClientV2) Get(key string) (expiringValue, bool) {
	c.mu.RLock()
	defer c.mu.Unlock()

	val, ok := c.data[key]
	return val, ok
}
