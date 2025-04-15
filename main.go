package main

import (
	"log"
	"net/http"
	"os"
	

	"github.com/sabrek15/clickhouse-ingestor/internal/auth"
	"github.com/sabrek15/clickhouse-ingestor/internal/storage"
	"github.com/sabrek15/clickhouse-ingestor/internal/filehandler"
	"github.com/sabrek15/clickhouse-ingestor/internal/api"
	"github.com/joho/godotenv"
	_ "github.com/ClickHouse/clickhouse-go"
)

func main() {

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Initialize auth with secret from env
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-please-change-me" // Fallback for development
	}
	jwtValidator := auth.NewJWTValidator(jwtSecret)

	// Initialize components
	clickhouseService := storage.NewClickHouseService()
	fileService := filehandler.NewFileService()
	apiHandler := api.NewAPIHandler(jwtValidator, clickhouseService, fileService)

	// Configure routes
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/connect", apiHandler.HandleConnect)
	mux.HandleFunc("/api/discover-schema", apiHandler.HandleSchemaDiscovery)
	mux.HandleFunc("/api/transfer", apiHandler.HandleDataTransfer)
	
	// UI routes
	mux.HandleFunc("/", apiHandler.HandleHome)
	mux.HandleFunc("/results", apiHandler.HandleResults)
	
	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}