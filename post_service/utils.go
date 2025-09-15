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
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		ServerPort:    os.Getenv("SERVER_PORT"),
		ServerHost:    os.Getenv("SERVER_HOST"),
		CachePassword: os.Getenv("CACHE_PASSWORD"),
		CacheHost:     os.Getenv("CACHE_HOST"),
		CachePort:     os.Getenv("CACHE_PORT"),
	}
	return config, nil
}

func InitDB(config models.Config) (*sql.DB, error) {
	DBpath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	DB, err := sql.Open("postgres", DBpath)
	if err != nil {
		log.Println("Failed to Connect with Post_service DB", err.Error())
		return nil, err
	}
	return DB, nil
}
