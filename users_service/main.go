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
	server := NewUserServer(UserRepo, config)

	log.Fatal(server.start())
}
