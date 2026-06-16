package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"undangan-digital/internal/handlers"
	"undangan-digital/internal/installer"
	"undangan-digital/internal/middleware"
	"undangan-digital/internal/repository"
	"undangan-digital/internal/routes"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runServer()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("undangan-digital version 1.0.0")
		return
	}

	ins := installer.NewInstaller()
	ins.Run()
}

func runServer() {
	_ = godotenv.Load()

	gin.SetMode(gin.ReleaseMode)

	dbPath := getEnv("DATABASE_PATH", "database/database.sqlite")
	port := getEnv("PORT", "8080")
	adminUser := getEnv("ADMIN_USER", "admin")
	adminPass := getEnv("ADMIN_PASS", "admin123")
	sessionSecret := getEnv("SESSION_SECRET", "default-secret-key-change-in-production")
	webhookURL := getEnv("BROADCAST_WEBHOOK_URL", "")
	appBaseURL := getEnv("APP_BASE_URL", "http://localhost:"+port)

	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.MkdirAll(dbDir, 0755)
	}

	repo, err := repository.NewRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	session := middleware.NewSessionManager(sessionSecret)
	broadcastService := handlers.NewBroadcastService("", "", appBaseURL)
	handler := handlers.NewHandler(repo, session, adminUser, adminPass, broadcastService)

	router := routes.SetupRouter(handler, session)
	router.SetTrustedProxies(nil)

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Admin panel: http://localhost:%s/admin/dashboard\n", port)
	fmt.Printf("Login: %s / %s\n", adminUser, adminPass)
	fmt.Printf("Database: %s\n", dbPath)
	if webhookURL != "" {
		fmt.Printf("Broadcast webhook: %s\n", webhookURL)
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}