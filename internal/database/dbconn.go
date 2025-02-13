package database

import (
	"Games/internal/config"
	"Games/internal/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var DB *gorm.DB

func getConnectionString(config *config.Config) string {
	dsn := fmt.Sprintf("host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBUserName, config.DBUserPassword, config.DBName)
	if config.DBPort != "" {
		dsn += fmt.Sprintf(" port=%v", config.DBPort)
	}
	return dsn
}

func InitDB(config *config.Config) {
	var err error
	dsn := getConnectionString(config)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the Database! \n", err.Error())
		os.Exit(1)
	}

	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Println("Running Migrations")
	err = DB.AutoMigrate(&models.User{}, &models.Game{}, &models.Skill{}, &models.UserSkill{})

	if err != nil {
		log.Fatal("Migration Failed:  \n", err.Error())
		os.Exit(1)
	}

	log.Println("🚀 Connected Successfully to the Database")
}
