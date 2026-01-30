package main

import (
	"category-management-api/database"
	"category-management-api/docs"
	"category-management-api/handlers"
	"category-management-api/repositories"
	"category-management-api/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Config holds application configuration
type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
	AppEnv string `mapstructure:"APP_ENV"`
}

// @title Category Management API
// @version 1.0
// @description RESTful API for managing categories and products with full CRUD operations
// @description
// @description ## Features:
// @description - Category Management (Get all, Get by ID, Create, Update, Delete)
// @description - Product Management (Get all with category names, Get by ID with category, Create, Update, Delete)
// @description - Product-Category Relationship (Foreign key with JOIN operations)
// @description
// @description ## Response Format:
// @description All endpoints return a standard response with:
// @description - status (bool): Request success status
// @description - message (string): Response message
// @description - data (object): Response data (when applicable)

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /
// @schemes http https

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API server is running
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Server is running"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "OK",
		"message": "Server is running successfully",
	})
}

// RootHandler shows API information
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        "Category Management API",
		"version":     "1.0",
		"status":      "running",
		"description": "RESTful API for managing categories and products",
		"endpoints": map[string]string{
			"documentation": "/docs/index.html",
			"health":        "/health",
			"categories":    "/categories",
			"products":      "/products",
		},
	})
}

// Global CORS middleware that wraps ALL handlers
type corsMiddleware struct {
	handler http.Handler
}

func (c *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Call the next handler
	c.handler.ServeHTTP(w, r)
}

func corsMiddlewareWrapper(handler http.Handler) http.Handler {
	return &corsMiddleware{handler: handler}
}

func main() {
	// Load configuration with Viper
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
		AppEnv: viper.GetString("APP_ENV"),
	}

	// Set default port
	if config.Port == "" {
		config.Port = "8080"
	}

	// Configure Swagger for production (HTTPS) or local (HTTP)
	if config.AppEnv == "production" {
		docs.SwaggerInfo.Host = "category-management-apis.zeabur.app"
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Host = "localhost:" + config.Port
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// ============================================
	// DATABASE CONNECTION
	// ============================================
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.CloseDB()

	// Run database migrations
	err = database.RunMigrations(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// ============================================
	// LAYERED ARCHITECTURE - DEPENDENCY INJECTION
	// ============================================

	// 1. Initialize Repository Layer (Data Access)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// 2. Initialize Service Layer (Business Logic)
	categoryService := services.NewCategoryService(categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo)

	// 3. Initialize Handler Layer (HTTP)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Root endpoint - API information
	mux.HandleFunc("/", RootHandler)

	// Health check endpoint
	mux.HandleFunc("/health", HealthCheck)

	// Category endpoints - using handler methods
	mux.HandleFunc("/categories", categoryHandler.HandleCategories)
	mux.HandleFunc("/categories/", categoryHandler.HandleCategoryByID)

	// Product endpoints - using handler methods
	mux.HandleFunc("/products", productHandler.HandleProducts)
	mux.HandleFunc("/products/", productHandler.HandleProductByID)

	// API Documentation endpoint
	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)

	// Start server with CORS middleware wrapping all routes
	addr := "0.0.0.0:" + config.Port
	fmt.Printf("Server running on %s\n", addr)
	fmt.Printf("API Documentation: http://localhost:%s/docs/index.html\n", config.Port)
	log.Println("Available endpoints:")
	log.Println("  GET    /health")
	log.Println("  GET    /categories")
	log.Println("  POST   /categories")
	log.Println("  GET    /categories/{id}")
	log.Println("  PUT    /categories/{id}")
	log.Println("  DELETE /categories/{id}")
	log.Println("  GET    /products")
	log.Println("  POST   /products")
	log.Println("  GET    /products/{id}")
	log.Println("  PUT    /products/{id}")
	log.Println("  DELETE /products/{id}")

	// Wrap the entire mux with CORS middleware
	handler := corsMiddlewareWrapper(mux)

	err = http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
