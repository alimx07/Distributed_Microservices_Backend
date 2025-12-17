package main

import (
	"encoding/base64"
	"log"
	"os"
	"strings"

	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
)

func ValidateCursorData(c string, pageSize int32) (string, int32) {
	decode, err := base64.StdEncoding.DecodeString(c)
	if err != nil || len(decode) == 0 {
		log.Println("invalid Cursor", err.Error())
		// continue with null cursor
		c = "-inf"
	}
	c = string(decode)
	if pageSize < 0 {
		log.Println("invalid PageSize")
		pageSize = 50 // default Value
	}

	return c, pageSize
}

func LoadConfig() (models.ServerConfig, error) {
	config := models.ServerConfig{
		ServerPort:     os.Getenv("SERVER_PORT"),
		ServerHost:     os.Getenv("SERVER_HOST"),
		ServerHTTPPort: os.Getenv("SERVER_HTTP_PORT"),
		PostService:    os.Getenv("POST_SERVICE"),
		UserService:    os.Getenv("USER_SERVICE"),
		FollowService:  os.Getenv("FOLLOW_SERVICE"),
	}
	return config, nil
}

func LoadRedisConfig() (models.RedisConfig, error) {
	config := models.RedisConfig{
		ClusterAddr: strings.Split(os.Getenv("CLUSTER_ADDR"), ","),
		Password:    os.Getenv("CACHE_PASSWORD"),
	}
	return config, nil
}

func LoadkakfaConfig() (models.KafkaConfig, error) {
	config := models.KafkaConfig{
		BootStrapServers: os.Getenv("BOOTSTRAP_SERVERS"),
		GroupID:          os.Getenv("GROUP_ID"),
		OffsetReset:      os.Getenv("OFFSET_RESET"),
		FetchMinBytes:    os.Getenv("FETCH_MIN_BYTES"),
		Topics:           strings.Split(os.Getenv("TOPICS"), ","),
	}
	return config, nil
}
