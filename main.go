package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"my-backend-go/config"
	"my-backend-go/handlers"
	"my-backend-go/middleware"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("[WARNING] .env file not found, using system environment")
	}

	// Pre-warm DB pool
	config.GetDB()

	// Setup Gin
	r := gin.Default()

	// ─── CORS Middleware ─────────────────────────────────────────────────────
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ─── Health Check ─────────────────────────────────────────────────────────
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Api E-testing v1. (Go + Gin)"})
	})

	// ─── /api/ ────────────────────────────────────────────────────────────────
	api := r.Group("/api")
	{
		// Pool management
		api.GET("/closepool", handlers.ClosePool)
	}

	// ─── /api/service ─────────────────────────────────────────────────────────
	svc := r.Group("/api/service")
	{
		// Authentication (no middleware required)
		svc.POST("/sign", middleware.SignToken)
		svc.POST("/verify", middleware.VerifyTokenEndpoint)

		// Microsoft login placeholder
		svc.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Microsoft login – implement OAuth flow here"})
		})

		// Test / utility endpoints
		svc.GET("/select", handlers.TestGetSelect)
		svc.GET("/testinsert/:x", handlers.TestInsert)
		svc.PUT("/send", handlers.TestRevData)
		svc.POST("/insertdb", handlers.Insertdb)
		svc.DELETE("/deletedb", handlers.Deletedb)
		svc.GET("/testpool", handlers.GetTestPool)
	}

	// ─── /api/service/frontend ────────────────────────────────────────────────
	fe := r.Group("/api/service/frontend")
	{
		// Authentication endpoints
		fe.POST("/sign", middleware.SignToken)
		fe.POST("/verify", middleware.VerifyTokenEndpoint)
	}

	// ─── Start Server ─────────────────────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("[Server] Listening on port %s\n", port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
