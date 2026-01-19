package main

import (
	"log"
	"os"

	"dev.sourcecraft.dolgintsev/golang-gin/internal/database"
	"dev.sourcecraft.dolgintsev/golang-gin/internal/handlers"
	"dev.sourcecraft.dolgintsev/golang-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Connect to database
	database.Connect()

	// Set Gin mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	// Create router
	r := gin.Default()

	// Prometheus metrics middleware
	r.Use(middleware.PrometheusMiddleware())

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		products := api.Group("/products")
		{
			products.GET("", handlers.GetAllProducts)
			products.GET("/:id", handlers.GetProductByID)
			products.POST("", handlers.CreateProduct)
			products.PUT("/:id", handlers.UpdateProduct)
			products.DELETE("/:id", handlers.DeleteProduct)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
