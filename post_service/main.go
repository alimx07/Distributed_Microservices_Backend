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
	cachedRepo, err := cachedRepo.NewRedisRepo(postRepo, config.CacheAddrs, config.CachePassword)
	if err != nil {
		log.Println("Error in Loading Redis Cluster: ", err.Error())
	}
	postService := NewPostService(postRepo, cachedRepo, config)
	go func() {
		log.Println(postService.StartHealthServer())
	}()
	log.Fatal(postService.start())
}
