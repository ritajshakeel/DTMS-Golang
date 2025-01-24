package config

import (
	"log"
	"sync"

	"dtms/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

func ConnectDatabase() {
	once.Do(func() {
		dsn := "dtms.db"
		var err error

		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		if err := DB.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
			log.Fatal("Failed to migrate database:", err)
		}

		log.Println("Database connected and migrated successfully")
	})
}
