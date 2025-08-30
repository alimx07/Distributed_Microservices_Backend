package main

import (
	"encoding/base64"
	"errors"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

func LoadConfig() (Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return Config{}, err
	}
	config := Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
		ServerHost: os.Getenv("SERVER_HOST"),
	}
	prv64 := os.Getenv("JWT_PRIVATE_KEY")
	prvBytes, err := base64.StdEncoding.DecodeString(prv64)
	if err != nil {
		return Config{}, errors.New("failed intialization of config")
	}
	config.JWTSecret = prvBytes
	return config, nil
}

// Mock Checking function
func check(user User) error {
	if len(user.UserName) == 0 {
		return errors.New("username cannot be empty")
	}
	if len(user.UserName) < 2 {
		return errors.New("username must be at least 3 characters")
	}
	if len(user.Email) == 0 {
		return errors.New("email cannot be empty")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		return errors.New("invalid email format")
	}
	if len(user.Password) < 4 {
		return errors.New("password must be at least 4 characters")
	}
	return nil
}
