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
	flowRepo := repository.NewFlowRepository(supabase)
	conversationRepo := repository.NewConversationRepository(supabase)
	wasapbotRepo := repository.NewWasapbotRepository(supabase)
	analyticsRepo := repository.NewAnalyticsRepository(supabase)
	stageRepo := repository.NewStageRepository(supabase)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	deviceService := service.NewDeviceService(deviceRepo)
	flowService := service.NewFlowService(flowRepo, deviceRepo)
	conversationService := service.NewConversationService(conversationRepo, deviceRepo)
	aiService := service.NewAIService(deviceRepo)
	whatsappService := service.NewWhatsAppService(deviceRepo)
	webhookService := service.NewWebhookService(deviceRepo, flowRepo)
	flowExecutionService := service.NewFlowExecutionService(flowRepo, conversationRepo, deviceRepo, aiService)
	flowProcessorService := service.NewFlowProcessorService(webhookService, whatsappService, flowRepo, deviceRepo, conversationRepo, wasapbotRepo, stageRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo, deviceRepo)
	stageService := service.NewStageService(stageRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(deviceService, authService)
	flowHandler := handler.NewFlowHandler(flowService, authService)
	conversationHandler := handler.NewConversationHandler(conversationService, authService)
	aiHandler := handler.NewAIHandler(aiService, authService)
	webhookHandler := handler.NewWebhookHandler(flowExecutionService, deviceService, whatsappService, flowProcessorService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, authService)
	stageHandler := handler.NewStageHandler(stageService)
	log.Printf("✅ Authentication system initialized")
	log.Printf("✅ Device management system initialized")
	log.Printf("✅ Flow builder system initialized")
	log.Printf("✅ Conversation management system initialized")
	log.Printf("✅ AI integration system initialized (OpenAI + Anthropic)")
	log.Printf("✅ WhatsApp messaging service initialized")
	log.Printf("✅ Flow execution engine initialized")
	log.Printf("✅ Analytics & reporting system initialized")
	log.Printf("✅ Stage value management system initialized")

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
	auth.Put("/change-password", authHandler.ChangePassword)
	auth.Put("/update-profile", authHandler.UpdateProfile)

	// Device management routes (requires authentication)
	devices := api.Group("/devices")
	devices.Post("/", deviceHandler.CreateDevice)
	devices.Get("/", deviceHandler.GetUserDevices)
	devices.Get("/:id", deviceHandler.GetDevice)
	devices.Put("/:id", deviceHandler.UpdateDevice)
	devices.Delete("/:id", deviceHandler.DeleteDevice)
	devices.Post("/:id/generate", deviceHandler.GenerateDevice)
	devices.Get("/:id/status", deviceHandler.CheckDeviceStatus)

	// Flow builder routes (requires authentication)
	flows := api.Group("/flows")
	flows.Post("/", flowHandler.CreateFlow)
	flows.Get("/", flowHandler.GetAllUserFlows)
	flows.Get("/device/:deviceId", flowHandler.GetFlowsByDevice)
	flows.Get("/:id", flowHandler.GetFlow)
	flows.Put("/:id", flowHandler.UpdateFlow)
	flows.Delete("/:id", flowHandler.DeleteFlow)

	// Conversation management routes (requires authentication)
	conversations := api.Group("/conversations")
	conversations.Get("/all", conversationHandler.GetAllConversations)
	conversations.Post("/", conversationHandler.CreateConversation)
	conversations.Get("/:id", conversationHandler.GetConversation)
	conversations.Put("/:id", conversationHandler.UpdateConversation)
	conversations.Delete("/:id", conversationHandler.DeleteConversation)
	conversations.Post("/:id/messages", conversationHandler.AddMessage)
	conversations.Get("/device/:deviceId", conversationHandler.GetConversationsByDevice)
	conversations.Get("/device/:deviceId/active", conversationHandler.GetActiveConversations)
	conversations.Get("/device/:deviceId/stats", conversationHandler.GetConversationStats)

	// AI integration routes (requires authentication)
	ai := api.Group("/ai")
	ai.Post("/completion", aiHandler.GenerateCompletion)
	ai.Post("/chat", aiHandler.SimpleChat)
	ai.Post("/test", aiHandler.TestConnection)

	// Webhook routes (public - no authentication required)
	webhook := api.Group("/webhook")
	webhook.Post("/:webhook_id", webhookHandler.ReceiveWebhook) // New unified webhook endpoint
	webhook.Post("/whatsapp/:deviceId", webhookHandler.HandleWhatsAppWebhook)
	webhook.Post("/waha/:deviceId", webhookHandler.HandleWahaWebhook)
	webhook.Post("/wablas/:deviceId", webhookHandler.HandleWablasWebhook)
	webhook.Post("/whacenter/:deviceId", webhookHandler.HandleWhacenterWebhook)
	webhook.Post("/start-flow", webhookHandler.StartFlow)

	// Analytics routes (requires authentication)
	analytics := api.Group("/analytics")
	analytics.Get("/dashboard", analyticsHandler.GetDashboard)
	analytics.Get("/conversations", analyticsHandler.GetConversationAnalytics)
	analytics.Get("/flows/:flowId", analyticsHandler.GetFlowAnalytics)
	analytics.Post("/export", analyticsHandler.ExportAnalytics)

	// Stage value routes (requires authentication)
	stages := api.Group("/stage-values")
	stages.Post("/", stageHandler.CreateStageValue)
	stages.Get("/", stageHandler.GetAllStageValues)
	stages.Get("/:id", stageHandler.GetStageValue)
	stages.Put("/:id", stageHandler.UpdateStageValue)
	stages.Delete("/:id", stageHandler.DeleteStageValue)

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
	app.Static("/assets", "./frontend/assets")
	app.Static("/", "./frontend", fiber.Static{
		Index: "index.html",
	})

	// Catch-all webhook route for custom URL patterns like /:userid/:flowname
	// This must be AFTER all /api routes to avoid conflicts
	app.Post("/:webhook_id/:flow_name", webhookHandler.ReceiveWebhook)

	// SPA fallback - serve index.html for all non-API routes
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		// Check if path starts with /api or is /healthz
		if len(path) >= 4 && path[:4] == "/api" {
			return c.Next()
		}
		if path == "/healthz" {
			return c.Next()
		}
		// Serve index.html for all other routes
		return c.SendFile("./frontend/index.html")
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
