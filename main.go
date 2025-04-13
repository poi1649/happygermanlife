package main

import (
	"awesomeProject2/handlers"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

func main() {
	// Command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Check for required environment variables
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Println("Warning: GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
		log.Println("Speech-to-Text functionality may not work correctly")
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Println("Warning: OPENAI_API_KEY environment variable is not set")
		log.Println("Response generation functionality will not work correctly")
	}

	// Initialize router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/api/speech", handlers.HandleSpeechToText)
	router.HandleFunc("/api/generate-response", handlers.HandleGenerateResponse).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Service is healthy")
	})

	// Configure CORS - allow everything for local testing
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"}, // Allow all headers
		ExposedHeaders:   []string{"*"}, // Expose all headers
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})

	// Create HTTP server
	handler := c.Handler(router)

	// Add middleware to add CORS headers to all responses for WebSocket as well
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers manually for WebSocket protocol upgrade requests
		if websocketRequest := r.Header.Get("Upgrade") == "websocket"; websocketRequest {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Handler(router).ServeHTTP(w, r)
	})

	// Start server
	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on %s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
