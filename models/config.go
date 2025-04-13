package models

import (
	"log"
	"os"
)

// GetOpenAIAPIKey returns the OpenAI API key from environment variables
func GetOpenAIAPIKey() string {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Println("Warning: OPENAI_API_KEY environment variable is not set")
	}
	return key
}

// GetGoogleCredentialsPath returns the path to Google Cloud credentials file
func GetGoogleCredentialsPath() string {
	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if path == "" {
		log.Println("Warning: GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}
	return path
}
