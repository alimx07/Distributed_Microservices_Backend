package main

type Config struct {
	JWTSecret []byte
	// JWTExpiration string

	ServerHost string
	ServerPort string
}
