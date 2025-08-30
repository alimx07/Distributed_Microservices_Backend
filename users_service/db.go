package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(config Config) *sql.DB {
	DBpath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	DB, err := sql.Open("postgres", DBpath)
	if err != nil {
		log.Fatal("Failed to Connect with DB", err.Error())
	}
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        password VARCHAR(100) NOT NULL
    );`)
	if err != nil {
		log.Fatal("Failed to create Users table: ", err.Error())
	}
	return DB
}
