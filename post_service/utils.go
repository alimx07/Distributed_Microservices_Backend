package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	_ "github.com/lib/pq"
)

func LoadConfig() (models.Config, error) {
	config := models.Config{
		// Primary DB
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		// Replica DB
		DBReplicaHost:     os.Getenv("DB_REPLICA_HOST"),
		DBReplicaPort:     os.Getenv("DB_REPLICA_PORT"),
		DBReplicaUser:     os.Getenv("DB_REPLICA_USER"),
		DBReplicaPassword: os.Getenv("DB_REPLICA_PASSWORD"),
		DBReplicaName:     os.Getenv("DB_REPLICA_NAME"),

		ServerPort:     os.Getenv("SERVER_PORT"),
		ServerHost:     os.Getenv("SERVER_HOST"),
		ServerHttpPort: os.Getenv("SERVER_HTTP_PORT"),
		CachePassword:  os.Getenv("CACHE_PASSWORD"),
		CacheHost:      os.Getenv("CACHE_HOST"),
		CachePort:      os.Getenv("CACHE_PORT"),
	}
	return config, nil
}

func InitDBConnections(config models.Config) (*sql.DB, *sql.DB, error) {
	// Primary connection (for writes)
	primaryPath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	primaryDB, err := sql.Open("postgres", primaryPath)
	if err != nil {
		log.Println("Failed to connect to primary DB:", err.Error())
		return nil, nil, err
	}

	// Replica connection (for reads)
	replicaPath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBReplicaHost, config.DBReplicaPort, config.DBReplicaUser,
		config.DBReplicaPassword, config.DBReplicaName)

	replicaDB, err := sql.Open("postgres", replicaPath)
	if err != nil {
		log.Println("Failed to connect to replica DB:", err.Error())
		return nil, nil, err
	}

	//connection pools
	// TODO:
	// Pool Numbers may need tunnig later
	primaryDB.SetMaxOpenConns(15)
	primaryDB.SetMaxIdleConns(5)

	replicaDB.SetMaxOpenConns(25)
	replicaDB.SetMaxIdleConns(10)

	return primaryDB, replicaDB, nil
}
