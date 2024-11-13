package database

import (
	"Games/internal/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getConnectionString(config *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUserName, config.DBUserPassword, config.DBName)
}

func InitDB(config *config.Config) error {
	connString := getConnectionString(config)
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&User{})

	return nil
}
