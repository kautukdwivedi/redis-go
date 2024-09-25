package main

import (
	"log"
)

func main() {
	config, err := newServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	server := newServer(config)

	go server.start()

	for {
		sleepSeconds(1)
	}
}
