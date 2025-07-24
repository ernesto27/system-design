package main

import (
	"log"
)

func main() {
	config := NewDatabaseConfig()

	db, err := ConnectDatabase(config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := SeedData(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	log.Println("Database connected, migrated, and seeded successfully!")
	
	router := SetupRouter(db)
	log.Println("Starting server on :8080")
	router.Run(":8080")
}
