package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitDB(config models.Config) (*sql.DB, error) {
	DBpath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	DB, err := sql.Open("postgres", DBpath)
	if err != nil {
		log.Println("Failed to Connect with Post_service DB", err.Error())
		return nil, err
	}
	err = applyMigration(DB, config.DBName)
	if err != nil {
		return nil, err
	}
	return DB, nil
}

func applyMigration(db *sql.DB, dbname string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	sourceURL := "github.com/alimx07/Distributed_Microservices_Backend/post_service/db"

	if err != nil {
		log.Println("Using Same Connection for Migrations failed :", err.Error())

	}
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		dbname,
		driver)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Migrations are appiled in transactions in postgress
	// So in case of fail , it will run rollback
	// No Need for down migrations scripts for now
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Println("Migration of Database failed: ", err.Error())
		return err
	}

	log.Println("Migrations applied successfully!")
	return nil
}
