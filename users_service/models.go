package main

type User struct {
	UserID   string `json:"userid,omitempty"`
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret []byte
	// JWTExpiration string

	ServerHost string
	ServerPort string
}
