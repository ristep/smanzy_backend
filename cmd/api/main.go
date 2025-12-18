package main

import (
	"fmt"
	"log"
	"os"

	// Gin is a web framework for Go (handling HTTP requests/responses)
	"github.com/gin-gonic/gin"
	// godotenv loads environment variables from a .env file
	"github.com/joho/godotenv"
	// GORM is an Object Relational Mapper (ORM) for database interactions
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// Internal packages from our own project
	"github.com/ristep/smanzy_backend/internal/auth"
	"github.com/ristep/smanzy_backend/internal/handlers"
	"github.com/ristep/smanzy_backend/internal/middleware"
	"github.com/ristep/smanzy_backend/internal/models"
)

// main is the entry point of the application
func main() {
	// 1. Load environment variables from .env file (if it exists)
	// This allows us to configure the app without changing code (e.g. secret keys, db passwords)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 2. Configuration Setup
	// We read configuration (Database URL, Secrets, Port) from the environment
	dbDSN := os.Getenv("DB_DSN") // Data Source Name (connection string)
	if dbDSN == "" {
		log.Fatal("DB_DSN environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET") // Secret key for signing JWT tokens
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT") // Port to run the server on
	if serverPort == "" {
		serverPort = "8080" // Default to 8080 if not specified
	}

	// 3. Database Connection
	// Connect to PostgreSQL using GORM
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 4. Database Migration
	// AutoMigrate automatically updates the database schema (tables) to display
	// the Go structs defined in `internal/models`.
	// Be careful with this in production!
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Media{}); err != nil {
		log.Fatalf("Failed to auto-migrate models: %v", err)
	}

	log.Println("Database migration completed successfully")

	// 5. Seeding Data
	// Ensure that basic roles exist in the database
	db.FirstOrCreate(&models.Role{}, models.Role{Name: "user"})
	db.FirstOrCreate(&models.Role{}, models.Role{Name: "admin"})

	// 6. Service Initialization
	// Initialize our services and handlers, injecting dependencies (like the DB connection)
	jwtService := auth.NewJWTService(jwtSecret)

	authHandler := handlers.NewAuthHandler(db, jwtService)
	userHandler := handlers.NewUserHandler(db)
	mediaHandler := handlers.NewMediaHandler(db)

	// 7. Router Setup
	// Create a new Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// Apply CORS middleware (Cross-Origin Resource Sharing) to allow frontend to talk to backend
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint - useful for monitoring if the app is up
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 8. Define Routes
	// Group routes under /api
	api := router.Group("/api")
	{
		// == PUBLIC ROUTES ==
		// These endpoints can be accessed without logging in

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.RegisterHandler)
			auth.POST("/login", authHandler.LoginHandler)
			auth.POST("/refresh", authHandler.RefreshHandler)
		}

		// Public media listing
		api.GET("/media", mediaHandler.ListPublicMediasHandler)

		// Serve uploaded files directly (for development)
		// :name is a path parameter that captures the filename
		api.GET("/media/files/:name", mediaHandler.ServeFileHandler)
	}

	// == PROTECTED ROUTES ==
	// Requires a valid JWT token in the Authorization header
	protectedAPI := router.Group("/api")
	// Apply the AuthMiddleware to check for the token
	protectedAPI.Use(middleware.AuthMiddleware(jwtService, db))
	{
		// Authenticated User routes
		profile := protectedAPI.Group("/profile")
		{
			profile.GET("", authHandler.ProfileHandler)       // Get current user profile
			profile.PUT("", authHandler.UpdateProfileHandler) // Update current user profile
		}

		// Admin-only routes
		// Apply RoleMiddleware to check if the user has "admin" role
		users := protectedAPI.Group("/users")
		users.Use(middleware.RoleMiddleware("admin"))
		{
			users.GET("", userHandler.GetAllUsersHandler)
			users.GET("/:id", userHandler.GetUserByIDHandler)
			users.PUT("/:id", userHandler.UpdateUserHandler)
			users.DELETE("/:id", userHandler.DeleteUserHandler)

			// Role management
			users.POST("/:id/roles", userHandler.AssignRoleHandler)
			users.DELETE("/:id/roles", userHandler.RemoveRoleHandler)
		}

		// Media routes (authenticated)
		media := protectedAPI.Group("/media")
		{
			media.POST("", mediaHandler.UploadHandler)            // Upload a new file
			media.GET("/:id", mediaHandler.GetMediaHandler)       // Get file details
			media.PUT("/:id", mediaHandler.UpdateMediaHandler)    // Edit file (Owner or Admin)
			media.DELETE("/:id", mediaHandler.DeleteMediaHandler) // Delete file (Owner or Admin)
		}

	}

	// 9. Start Server
	addr := fmt.Sprintf(":%s", serverPort)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
