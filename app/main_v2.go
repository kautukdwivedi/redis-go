package main

import (
	"log"
	"log/slog"
)

func main() {
	config, err := serverConfig()
	if err != nil {
		slog.Error("server config error", "err", err)
	}

	server := NewServer(config)
	log.Fatal(server.Start())
}
