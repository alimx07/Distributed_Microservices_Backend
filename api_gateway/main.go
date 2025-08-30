package main

import "log"

func main() {
	InitLogger()
	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	handler := NewApiHandler()
	server := NewServer(handler, config)
	log.Fatal(server.start())
}
