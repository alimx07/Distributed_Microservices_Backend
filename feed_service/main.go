package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

	errChan := make(chan error, 1)

	go func() {
		if err := service.StartHealthServer(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	go func() {
		if err := service.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("ShutDown Signal received")
	case err := <-errChan:
		log.Println("Service Error: ", err.Error())
	}

	stop()

	service.close()
}
