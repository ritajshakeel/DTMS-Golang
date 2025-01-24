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

// ConnectDatabase initializes the database connection and runs migrations.
func ConnectDatabase() {
	once.Do(func() {
		dsn := "dtms.db" // SQLite database file
		var err error

		// Open the database connection
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// AutoMigrate models: Automatically create tables, add missing columns, and indexes
		if err := DB.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
			log.Fatal("Failed to migrate database:", err)
		}

		log.Println("Database connected and migrated successfully")
	})
}
