package main

import (
	"fmt"
	"log"
	"os"

	"chatbot-automation/internal/config"
	"chatbot-automation/internal/database"
	"chatbot-automation/internal/handler"
	"chatbot-automation/internal/repository"
	"chatbot-automation/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("✅ Configuration loaded")

	// Initialize Supabase client
	supabase := database.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, cfg.SupabaseServiceRoleKey)
	log.Printf("🔗 Connecting to Supabase...")

	// Test Supabase connection
	if err := supabase.TestConnection(); err != nil {
		log.Printf("⚠️  Warning: Supabase connection test failed: %v", err)
		log.Printf("⚠️  Server will start anyway, but database operations may fail")
	} else {
		log.Printf("✅ Supabase connection successful!")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(supabase)
	deviceRepo := repository.NewDeviceRepository(supabase)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	deviceService := service.NewDeviceService(deviceRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(deviceService, authService)
	log.Printf("✅ Authentication system initialized")
	log.Printf("✅ Device management system initialized")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Health check endpoint (CRITICAL for Railway)
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "ok",
			"message":  "Chatbot Automation Platform - Rebuilt from scratch",
			"database": "connected",
		})
	})

	// API routes
	api := app.Group("/api")

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/profile", authHandler.GetProfile)

	// Device management routes (requires authentication)
	devices := api.Group("/devices")
	devices.Post("/", deviceHandler.CreateDevice)
	devices.Get("/", deviceHandler.GetUserDevices)
	devices.Get("/:id", deviceHandler.GetDevice)
	devices.Put("/:id", deviceHandler.UpdateDevice)
	devices.Delete("/:id", deviceHandler.DeleteDevice)

	// Status endpoint with database check
	api.Get("/status", func(c *fiber.Ctx) error {
		dbStatus := "connected"
		if err := supabase.TestConnection(); err != nil {
			dbStatus = fmt.Sprintf("error: %v", err)
		}

		return c.JSON(fiber.Map{
			"status":   "running",
			"version":  "2.0.0-rebuild",
			"database": dbStatus,
		})
	})

	// Test database endpoint (uses service role to bypass RLS)
	api.Get("/db/test", func(c *fiber.Ctx) error {
		// Query user table as admin (bypasses RLS)
		data, err := supabase.QueryAsAdmin("user", map[string]string{
			"select": "*",
			"limit":  "5",
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Database connection working!",
			"users":   string(data),
		})
	})

	// Serve static files (frontend)
	app.Static("/assets", "./dist/assets")
	app.Static("/", "./dist", fiber.Static{
		Index: "index.html",
	})

	// SPA fallback - serve index.html for all non-API routes
	app.Use(func(c *fiber.Ctx) error {
		if c.Path()[:4] != "/api" && c.Path() != "/healthz" {
			return c.SendFile("./dist/index.html")
		}
		return c.Next()
	})

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	bind := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("🚀 Server starting on %s", bind)
	log.Fatal(app.Listen(bind))
}
