package main

import (
	"fmt"
	"log"
	"net/http"
	"retail-core-api/config"
	"retail-core-api/database"
	"retail-core-api/docs"
	"retail-core-api/handlers"
	"retail-core-api/helpers"
	"retail-core-api/middleware"
	"retail-core-api/repositories"
	"retail-core-api/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Retail Core API
// @version 1.0
// @description RESTful API for managing categories, products, transactions, and POS operations
// @description
// @description ## Features:
// @description - Authentication (JWT login/register)
// @description - User Management (owner-only)
// @description - Category Management (CRUD)
// @description - Product Management (CRUD with category, search, pagination)
// @description - Transaction / Checkout (multi-item with payment method, discount, notes)
// @description - Void Transactions
// @description - Sales Reports (daily, date range, summary with category breakdown)
// @description - Dashboard Statistics

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Configure Swagger
	docs.SwaggerInfo.Host = cfg.SwaggerHost()
	docs.SwaggerInfo.Schemes = cfg.SwaggerSchemes()

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// ============================================
	// DATABASE CONNECTION
	// ============================================
	db, err := database.InitDB(cfg.DBConn)
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
	// DEPENDENCY INJECTION
	// ============================================

	// Repositories
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Services
	categoryService := services.NewCategoryService(categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo)
	transactionService := services.NewTransactionService(transactionRepo)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo)

	// Handlers
	categoryHandler := handlers.NewCategoryHandler(categoryService, productService)
	productHandler := handlers.NewProductHandler(productService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// ============================================
	// ROUTER SETUP
	// ============================================
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// ── Health & Info ──────────────────────────
	r.GET("/health", func(c *gin.Context) {
		helpers.OK(c, "Server is running successfully", gin.H{"status": "OK"})
	})

	r.GET("/", func(c *gin.Context) {
		helpers.OK(c, "Retail Core API", gin.H{
			"name":    "Retail Core API",
			"version": "1.0",
			"status":  "running",
		})
	})

	// ── Swagger Documentation ─────────────────
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ── Auth (public) ─────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}

	// ── Protected API routes ──────────────────
	api := r.Group("/api")
	api.Use(middleware.Auth(cfg.JWTSecret))
	{
		// Categories
		api.GET("/categories", categoryHandler.List)
		api.GET("/categories/:id", categoryHandler.GetByID)
		api.GET("/categories/:id/products", categoryHandler.GetProducts)
		api.POST("/categories", categoryHandler.Create)
		api.PUT("/categories/:id", categoryHandler.Update)
		api.DELETE("/categories/:id", categoryHandler.Delete)

		// Products
		api.GET("/products", productHandler.List)
		api.GET("/products/:id", productHandler.GetByID)
		api.POST("/products", productHandler.Create)
		api.PUT("/products/:id", productHandler.Update)
		api.DELETE("/products/:id", productHandler.Delete)

		// Transactions / Checkout
		api.POST("/checkout", transactionHandler.Checkout)
		api.GET("/transactions", transactionHandler.ListTransactions)
		api.GET("/transactions/:id", transactionHandler.GetTransactionByID)
		api.PATCH("/transactions/:id/void", transactionHandler.VoidTransaction)

		// Dashboard
		api.GET("/dashboard", transactionHandler.Dashboard)

		// Reports
		api.GET("/report/today", transactionHandler.DailyReport)
		api.GET("/report", transactionHandler.ReportByRange)
		api.GET("/report/summary", transactionHandler.ReportSummary)

		// Users (owner only)
		users := api.Group("/users")
		users.Use(middleware.RequireRole("owner"))
		{
			users.GET("", userHandler.GetAll)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}

	// ── Start Server ──────────────────────────
	addr := "0.0.0.0:" + cfg.Port
	fmt.Printf("Server running on %s\n", addr)
	fmt.Printf("API Documentation: http://localhost:%s/docs/index.html\n", cfg.Port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
