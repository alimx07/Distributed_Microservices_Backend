package main

import (
	"encoding/base64"
	"errors"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func LoadConfig() (Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return Config{}, err
	}
	config := Config{
		ServerPort: os.Getenv("SERVER_PORT"),
		ServerHost: os.Getenv("SERVER_HOST"),
	}
	pubB64 := os.Getenv("JWT_PUBLIC_KEY")
	pubBytes, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil {
		return Config{}, errors.New("failed intialization of config")
	}
	config.JWTSecret = pubBytes
	return config, nil
}

func InitLogger() {
	f, _ := os.OpenFile("api_gateway.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func ValidateToken(token string, config Config) (string, error) {

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithAudience("api_gateway"),
		jwt.WithIssuer("users_service"),
	)
	claims := jwt.RegisteredClaims{}
	parse, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errors.New("token expired")
		}
		return "", err
	}
	if !parse.Valid {
		return "", errors.New("token is not valid")
	}
	if claims.Subject == "" {
		return "", errors.New("token is not valid")
	}
	return claims.Subject, nil
}
