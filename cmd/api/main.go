package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ristep/smanzy_backend/internal/auth"
	"github.com/ristep/smanzy_backend/internal/handlers"
	"github.com/ristep/smanzy_backend/internal/middleware"
	"github.com/ristep/smanzy_backend/internal/models"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get configuration from environment
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		log.Fatal("DB_DSN environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Media{}); err != nil {
		log.Fatalf("Failed to auto-migrate models: %v", err)
	}

	log.Println("Database migration completed successfully")

	// Create default roles if they don't exist
	db.FirstOrCreate(&models.Role{}, models.Role{Name: "user"})
	db.FirstOrCreate(&models.Role{}, models.Role{Name: "admin"})

	// Initialize JWT service
	jwtService := auth.NewJWTService(jwtSecret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtService)
	userHandler := handlers.NewUserHandler(db)
	mediaHandler := handlers.NewMediaHandler(db)

	// Create Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	api := router.Group("/api")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.RegisterHandler)
			auth.POST("/login", authHandler.LoginHandler)
			auth.POST("/refresh", authHandler.RefreshHandler)
		}
	}

	// Protected routes (requires authentication)
	protectedAPI := router.Group("/api")
	protectedAPI.Use(middleware.AuthMiddleware(jwtService, db))
	{
		// User profile routes
		profile := protectedAPI.Group("/profile")
		{
			profile.GET("", authHandler.ProfileHandler)
		}

		// User management routes (admin only)
		users := protectedAPI.Group("/users")
		users.Use(middleware.RoleMiddleware("admin"))
		{
			users.GET("", userHandler.GetAllUsersHandler)
			users.GET("/:id", userHandler.GetUserByIDHandler)
			users.PUT("/:id", userHandler.UpdateUserHandler)
			users.DELETE("/:id", userHandler.DeleteUserHandler)
			users.POST("/:id/roles", userHandler.AssignRoleHandler)
			users.DELETE("/:id/roles", userHandler.RemoveRoleHandler)
		}

		// Media routes
		media := protectedAPI.Group("/media")
		{
			media.POST("", mediaHandler.UploadHandler)            // Upload
			media.GET("/:id", mediaHandler.GetMediaHandler)       // Download (Read-only for others)
			media.PUT("/:id", mediaHandler.UpdateMediaHandler)    // Edit (Owner/Admin)
			media.DELETE("/:id", mediaHandler.DeleteMediaHandler) // Delete (Owner/Admin)
		}
	}

	// Start server
	addr := fmt.Sprintf(":%s", serverPort)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
