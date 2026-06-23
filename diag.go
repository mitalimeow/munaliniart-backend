package main

import (
	"context"
	"log"
	"go_tutorials/internal/config"
	"go_tutorials/internal/database"
)

func main() {
	appConfig := config.LoadConfig()
	dbPool, err := database.ConnectDB(appConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer dbPool.Close()

	query := `INSERT INTO enquiries (sender_name, sender_email, message) VALUES ($1, $2, $3)`
	_, err = dbPool.Exec(context.Background(), query, "test", "test@test.com", "test msg")
	if err != nil {
		log.Fatalf("Insert error: %v", err)
	}
	log.Println("Insert success")
}
