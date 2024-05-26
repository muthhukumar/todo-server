package internal

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadDotEnvFile() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
