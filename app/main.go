package main

import (
	"fmt"
	"log"
)

func main() {
	config, err := newServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	server := newServer(config)

	_, err = server.loadRDB()
	if err != nil {
		fmt.Println("error loading rdb file: ", err.Error())
	}

	go server.start()

	for {
		sleepSeconds(1)
	}
}
