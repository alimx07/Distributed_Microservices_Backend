package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appConfig, err := LoadConfig()
	if err != nil {
		log.Fatal("Error in Loading Service Config: ", err.Error())
	}
	KafkaConfig, err := LoadkakfaConfig()
	if err != nil {
		log.Fatal("Error in Loading Kafka Config: ", err.Error())
	}
	cacheConfig, err := LoadRedisConfig()
	if err != nil {
		log.Fatal("Error in Loading Redis Config: ", err.Error())
	}

	service, err := NewFeedService(appConfig, KafkaConfig, cacheConfig)
	if err != nil {
		log.Fatal("Error in Starting Feed Service: ", err.Error())
	}
	go func() {
		service.StartHealthServer()
	}()
	go func() {
		log.Fatal(service.Start())
	}()

	<-ctx.Done()

	service.close()
}
