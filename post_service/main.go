package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config file: ", err.Error())
	}

	primaryDB, replicaDB, err := InitDBConnections(config)
	if err != nil {
		log.Fatal("Failed to initialize database connections: ", err.Error())
	}

	postRepo := postRepo.NewPostgresRepo(primaryDB, replicaDB)
	cachedRepo, err := cachedRepo.NewRedisRepo(postRepo, config.CacheAddrs, config.CachePassword)
	if err != nil {
		log.Println("Error in Loading Redis Cluster: ", err.Error())
	}
	postService := NewPostService(postRepo, cachedRepo, config)

	errChan := make(chan error, 1)

	go func() {
		if err := postService.StartHealthServer(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	go func() {
		if err := postService.start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-errChan:
		log.Printf("Server error: %v", err)
	}

	stop()

	postService.close()
}
