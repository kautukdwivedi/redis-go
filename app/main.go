package main

import (
	"log"
)

func main() {
	err := serverConfig()
	if err != nil {
		log.Fatal(err)
	}

	go startServer()

	for {
		sleepSeconds(1)
	}
}
