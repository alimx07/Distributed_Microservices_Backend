package main

import (
	"log"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config file: ", err.Error())
	}

	primaryDB, replicaDB, err := InitDBConnections(config)
	if err != nil {
		log.Fatal("Failed to initialize database connections: ", err.Error())
	}
	defer primaryDB.Close()
	defer replicaDB.Close()

	postRepo := postRepo.NewPostgresRepo(primaryDB, replicaDB)
	cachedRepo := cachedRepo.NewRedisRepo(postRepo, config.CacheHost, config.CachePort, config.CachePassword)
	postService := NewPostService(postRepo, cachedRepo, config)
	log.Fatal(postService.start())
}
