package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contract-analysis-service/configs"
	_ "contract-analysis-service/docs" // Import docs for swagger
	swagger "contract-analysis-service/docs"
	"contract-analysis-service/internal/pkg/container"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	swagger.SwaggerInfo.Title = "Smart DocParser API"
	swagger.SwaggerInfo.Description = "API for document analysis and processing"
	swagger.SwaggerInfo.Version = "1.0"
	swagger.SwaggerInfo.Host = "localhost:" + port
	swagger.SwaggerInfo.BasePath = "/"
	swagger.SwaggerInfo.Schemes = []string{"http"}

	r := gin.Default()

	// Register middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Add CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	r.Use(cors.New(config))

	// Handlers
	healthHandler := c.NewHealthHandler()
	docHandler := c.NewDocumentHandler()
	analysisHandler := c.NewAnalysisHandler()
	contractAnalysisHandler := c.NewContractAnalysisHandler()

	// Health routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// Document routes
	r.POST("/contracts/upload", docHandler.Upload)
	r.POST("/contracts/upload-analyze", docHandler.UploadAndAnalyze)
	r.GET("/contracts/:id", docHandler.Get)
	r.DELETE("/contracts/:id", docHandler.Delete)

	// Analysis routes
	r.POST("/contracts/analyze", analysisHandler.Analyze)
	
	// Contract Analysis routes (Task 5.1)
	r.POST("/contracts/analyze-comprehensive", contractAnalysisHandler.AnalyzeContractComprehensively)
	r.GET("/contract-analysis/:contract_id", contractAnalysisHandler.GetContractAnalysis)
	r.GET("/contract-analyses", contractAnalysisHandler.ListContractAnalyses)
	r.GET("/analysis/:id", contractAnalysisHandler.GetAnalysisById)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Kill (no param) default send syscall.SIGTERM
	// Kill -2 is syscall.SIGINT
	// Kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
