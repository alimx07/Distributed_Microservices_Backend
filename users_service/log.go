package main

import (
	"log"
	"os"
)

func InitLogger() {
	f, _ := os.OpenFile("users_service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
