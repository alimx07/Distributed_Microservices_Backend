package main

import (
	"log"
)

func main() {
	InitLogger()
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed To Load The Configuration:", err.Error())
	}
	db := InitDB(config)
	UserRepo := NewUserRepo(db)
	server := NewUserServer(UserRepo, *config)

	log.Printf("Users service is running on %s:%s...", config.ServerHost, config.ServerPort)
	log.Fatal(server.start())
}
