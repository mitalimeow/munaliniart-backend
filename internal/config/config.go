package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	PasswordOne string
	PasswordTwo string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	passOne := os.Getenv("ADMIN_PASSWORD_ONE")
	passTwo := os.Getenv("ADMIN_PASSWORD_TWO")

	if dbURL == "" || jwtSecret == "" || passOne == "" || passTwo == "" {
		log.Fatal("CRITICAL ERROR: Missing vital environment variables in configuration profile.")
	}

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return &Config{
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		ServerPort:  serverPort,
		PasswordOne: passOne,
		PasswordTwo: passTwo,
	}
}
