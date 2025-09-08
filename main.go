package main

import (
	"log"

	"contract-analysis-service/docs"
	"contract-analysis-service/configs"
	"contract-analysis-service/internal/pkg/container"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()
	// Load config and initialize container
	cfg, err := configs.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	c := container.NewContainer(cfg)

	// Determine port from config (fallback to 8080)
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// Initialize Swagger (after determining port)
	docs.SwaggerInfo.Title = "Smart DocParser API"
	docs.SwaggerInfo.Description = "API for document analysis and processing"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	r := gin.Default()

	// Register middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Handlers
	healthHandler := c.NewHealthHandler()
	docHandler := c.NewDocumentHandler()

	// Health routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// Document routes
	r.POST("/contracts/upload", docHandler.Upload)
	r.GET("/contracts/:id", docHandler.Get)
	r.DELETE("/contracts/:id", docHandler.Delete)

	// Swagger route with error logging
	r.GET("/swagger/*any", func(c *gin.Context) {
		log.Printf("Accessing Swagger: %s", c.Request.URL.Path)
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		if c.Writer.Status() >= 400 {
			log.Printf("Swagger error: %d - %s", c.Writer.Status(), c.Errors.String())
		}
	})

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
