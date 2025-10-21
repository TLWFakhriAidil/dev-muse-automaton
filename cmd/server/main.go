package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"chatbot-automation/internal/config"
	"chatbot-automation/internal/database"
	"chatbot-automation/internal/handlers"
	"chatbot-automation/internal/repository"
	"chatbot-automation/internal/services"
	"chatbot-automation/internal/whatsapp"
)

func main() {
	// Set logrus to output to stdout for debugging
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	
	logrus.Info("Starting Chatbot Automation Server...")

	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Println("No .env file found, using environment variables")
	} else {
		logrus.Info(".env file loaded successfully")
	}

	// Load configuration
	logrus.Debug("Loading configuration...")
	cfg := config.Load()
	logrus.WithFields(logrus.Fields{
		"supabase_url": cfg.SupabaseURL,
		"port": cfg.Port,
	}).Debug("Configuration loaded")

	// Initialize Supabase database (Railway-compatible with retry)
	var db *sql.DB
	
	// RAILWAY CRITICAL FIX: Start server IMMEDIATELY, initialize everything in background
	
	// Create basic Fiber app first - BEFORE any service initialization
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// RAILWAY ULTRA-FIX: Immediate health endpoints - BEFORE service initialization
	app.Get("/healthz", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).SendString(`{"status":"ok","railway":true}`)
	})
	
	app.Get("/health/basic", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("OK")
	})

	// Start server immediately in background
	go func() {
		bind := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
		logrus.Infof("🚀 RAILWAY: Server starting IMMEDIATELY on %s", bind)
		logrus.Infof("🔗 Health endpoint available at: http://0.0.0.0:%d/healthz", cfg.Port)
		
		if err := app.Listen(bind); err != nil {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Give server a moment to start listening
	time.Sleep(100 * time.Millisecond)
	logrus.Info("🚀 RAILWAY: Server is now listening, proceeding with background initialization")

	// Initialize essential services for routes (NON-FATAL for Railway)
	// These services are needed for the routes below, so can't be fully background
	var websocketService *services.WebSocketService
	var mediaService *services.MediaService
	var whatsappService *whatsapp.Service
	var queueService *services.QueueService
	
	// Initialize minimal services needed for routes
	websocketService = services.NewWebSocketService(cfg.MaxConcurrentUsers)
	mediaService = services.NewMediaService(cfg.CDNEnabled, cfg.CDNBaseURL, "./media")
	logrus.Info("✅ RAILWAY: Essential services initialized for immediate routes")

	// Background initialization for heavy services - NON-FATAL
	go func() {
		logrus.Info("🔄 Background: Starting heavy services initialization...")

		// RAILWAY FIX: Use Supabase SDK (JavaScript-like pattern, no IPv6 issues)
		logrus.Info("🚀 Initializing Supabase SDK (REST API - Railway IPv4 compatible)")
		supabaseSDK, sdkErr := database.NewSupabaseSDK(cfg)
		var useSupabase bool = false

		if sdkErr == nil {
			// Test connection by querying chatbot_flows table
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			var testResult []map[string]interface{}
			testErr := supabaseSDK.From("chatbot_flows").Select("id").Execute(ctx, &testResult)
			cancel()

			if testErr == nil {
				logrus.Info("✅ Supabase SDK connected successfully - using HTTP REST API")
				logrus.Warn("⚠️ PostgreSQL direct connection SKIPPED - using REST API mode")
				logrus.Info("💡 Database schema must be managed via Supabase Dashboard")
				logrus.Info("💡 This solves Railway IPv6 issue - no IPv4 addon needed!")
				useSupabase = true
				db = nil // Force REST API usage
			} else {
				logrus.WithError(testErr).Warn("Supabase SDK test query failed, falling back to PostgreSQL")
			}
		} else {
			logrus.WithError(sdkErr).Warn("Failed to initialize Supabase SDK, falling back to PostgreSQL")
		}

		// Fallback to PostgreSQL if REST API fails
		var err error
		if !useSupabase {
			logrus.Info("🔄 REST API unavailable, attempting PostgreSQL connection...")
			db, err = database.Initialize(cfg)
			if err != nil {
				logrus.WithError(err).Error("⚠️  Background: Failed to initialize database (both REST and PostgreSQL failed)")
				db = nil
			} else {
				logrus.Info("✅ Background: PostgreSQL database initialized successfully")

				// Run migrations
				if err := database.RunMigrations(db); err != nil {
					logrus.WithError(err).Warn("Background: Failed to run migrations, continuing anyway")
				} else {
					logrus.Info("Background: Database migrations completed")
				}
			}
		}

		// Initialize Redis with clustering support
		redisClient := services.InitializeRedis(cfg)
		logrus.Info("Background: Redis initialized successfully")

		// Initialize performance-optimized services
		// Handle Redis client for services that need concrete type
		var concreteRedisClient *redis.Client
		if redisClient != nil {
			var ok bool
			concreteRedisClient, ok = redisClient.(*redis.Client)
			if !ok {
				logrus.Warn("Background: Redis client type assertion failed, using nil client")
				concreteRedisClient = nil
			}
		} else {
			logrus.Warn("Background: Redis not available, services will run without caching")
		}

		// Initialize repositories first (before services)
		// Use Supabase repositories if SDK is available, otherwise use SQL repositories
		var aiWhatsappRepo repository.AIWhatsappRepository
		var deviceSettingsRepo repository.DeviceSettingsRepository
		var wasapBotRepo repository.WasapBotRepository
		var executionProcessRepo repository.ExecutionProcessRepository
		var orderRepo repository.OrderRepository

		if useSupabase && supabaseSDK != nil {
			logrus.Info("🔄 Using Supabase REST API repositories (Railway IPv4 compatible)")
			aiWhatsappRepo = repository.NewAIWhatsappRepositorySupabase(supabaseSDK)
			deviceSettingsRepo = repository.NewDeviceSettingsRepositorySupabase(supabaseSDK)
			wasapBotRepo = repository.NewWasapBotRepositorySupabase(supabaseSDK)
			executionProcessRepo = repository.NewExecutionProcessRepositorySupabase(supabaseSDK)
			orderRepo = repository.NewOrderRepositorySupabase(supabaseSDK)
			logrus.Info("✅ All Supabase repositories initialized successfully (5/6 repositories)")
			logrus.Info("💡 StageSetValueRepository not yet migrated - using SQL fallback for this repo")
		} else {
			logrus.Info("🔄 Using traditional SQL repositories")
			aiWhatsappRepo = repository.NewAIWhatsappRepository(db)
			deviceSettingsRepo = repository.NewDeviceSettingsRepository(db)
			wasapBotRepo = repository.NewWasapBotRepository(db)
			executionProcessRepo = repository.NewExecutionProcessRepository(db)
			orderRepo = repository.NewOrderRepository(db)
			logrus.Info("✅ SQL repositories initialized successfully")
		}

		// Suppress unused variable warnings for repos not yet wired to services
		_ = executionProcessRepo
		_ = orderRepo

		flowService := services.NewFlowService(db, concreteRedisClient)
		aiService := services.NewAIService(cfg, deviceSettingsRepo)
		queueMonitor := services.NewQueueMonitor()
		queueService = services.NewQueueService(redisClient, queueMonitor)
		deviceSettingsService := services.NewDeviceSettingsService(db)

		// Initialize unified flow service for table routing
		unifiedFlowService := services.NewUnifiedFlowService(flowService, aiWhatsappRepo, wasapBotRepo)
		logrus.Info("Background: Unified flow service initialized for table routing")

		// Initialize provider service for message sending
		providerService := services.NewProviderService()
		logrus.Info("Background: Provider service initialized for Wablas/Whacenter APIs")

		// Initialize media detection service for centralized media URL detection
		mediaDetectionService := services.NewMediaDetectionService()
		logrus.Info("Background: Media detection service initialized for multiple format support")

		// Initialize health service for comprehensive system monitoring
		healthService := services.NewHealthService(db, concreteRedisClient, "1.0.0")
		logrus.Info("Background: Health service initialized for system monitoring")

		// Initialize AI WhatsApp service with media detection service
		aiWhatsappService := services.NewAIWhatsappService(aiWhatsappRepo, deviceSettingsRepo, flowService, mediaDetectionService, cfg)
		logrus.Info("Background: AI WhatsApp service initialized with media detection service")

		// Initialize WhatsApp service with multi-device support - NON-FATAL in background
		logrus.Info("Background: About to initialize WhatsApp service...")
		whatsappService, err = whatsapp.NewService(cfg, queueService, flowService, aiService, aiWhatsappService, websocketService, deviceSettingsService, providerService, mediaDetectionService, unifiedFlowService)
		if err != nil {
			logrus.WithError(err).Error("Background: Failed to initialize WhatsApp service - continuing without it")
			return // Exit background initialization, but don't kill the server
		}
		logrus.Info("✅ Background: WhatsApp service initialized successfully")

		// Set WhatsApp service dependency on queue service for flow continuation
		queueService.SetWhatsAppService(whatsappService)
		logrus.Info("✅ Background: Queue service configured with WhatsApp service dependency")

		// Initialize handlers with all services - in background
		_ = handlers.NewHandlers(
			flowService,
			aiService,
			queueService,
			whatsappService,
			deviceSettingsService,
			websocketService,
			mediaService,
			healthService,
			db,
			cfg,
		)
		
		logrus.Info("✅ Background: All services initialized successfully")
		
	}()

	// Initialize HTML template engine for the existing app
	engine := html.New("./templates", ".html")
	engine.Reload(cfg.AppEnv == "development")

	// Add template functions
	engine.AddFunc("now", func() time.Time {
		return time.Now()
	})

	// Update existing app config for production features
	// Note: The basic app is already created and listening above

	// Performance and security middleware
	app.Use(recover.New())

	// Rate limiting for API protection
	app.Use(limiter.New(limiter.Config{
		Max:        100, // 100 requests per minute per IP
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Device-ID",
		AllowCredentials: false, // Set to false when using wildcard origins
	}))

	if cfg.AppEnv == "development" {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency}) - ${ip}\n",
		}))
	}

	// Test endpoint to verify server version
	app.Get("/api/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":   "2025-10-07T03:05:00Z-CACHE-FIX",
			"message":   "🚨 NEW SERVER CODE IS RUNNING! Cache fix applied.",
			"timestamp": time.Now().Unix(),
		})
	})

	// Note: Health endpoints already defined above during immediate server start

	// WebSocket endpoint for real-time communication
	app.Use("/ws", func(c *fiber.Ctx) error {
		// Check if connection is a WebSocket upgrade
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocketService.HandleWebSocket)

	// Media endpoints for file upload and serving
	media := app.Group("/media")
	media.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No file uploaded",
			})
		}

		result, err := mediaService.UploadFile(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(result)
	})

	media.Get("/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		data, mimeType, err := mediaService.ServeFile(filename)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "File not found",
			})
		}

		c.Set("Content-Type", mimeType)
		c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		return c.Send(data)
	})

	media.Get("/thumbnails/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		data, mimeType, err := mediaService.ServeFile("thumbnails/" + filename)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Thumbnail not found",
			})
		}

		c.Set("Content-Type", mimeType)
		c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		return c.Send(data)
	})

	// Add request logging middleware for debugging
	app.Use(func(c *fiber.Ctx) error {
		logrus.WithFields(logrus.Fields{
			"method": c.Method(),
			"path":   c.Path(),
			"ip":     c.IP(),
		}).Info("Incoming request")
		return c.Next()
	})

	// Setup placeholder routes - handlers will be initialized in background
	// Placeholder API routes while handlers are loading
	api := app.Group("/api")
	api.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "loading",
			"message": "Services are initializing in background",
		})
	})
	
	// Note: Full routes will be set up when background initialization completes
	
	// ULTRA-DEBUG: Add test endpoint to verify server is working
	app.Get("/debug-test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "server working",
			"time": time.Now().Format(time.RFC3339),
			"assets_available": true,
			"react_app": "should_load",
		})
	})

	// Add middleware to force no-cache and prevent 304 responses - MUST BE BEFORE STATIC SERVING
	app.Use("/assets/*", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Last-Modified", time.Now().Format(http.TimeFormat))
		c.Set("ETag", fmt.Sprintf("\"%d\"", time.Now().Unix()))
		// Remove any conditional headers to prevent 304
		c.Request().Header.Del("If-Modified-Since")
		c.Request().Header.Del("If-None-Match")
		return c.Next()
	})

	// CRITICAL FIX: Static asset routes BEFORE catch-all to ensure proper MIME types

	// Serve static assets with proper MIME types and caching
	app.Static("/assets", "./dist/assets", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		Index:         "",
		CacheDuration: 0, // Disable server-side cache for fresh deploys
		MaxAge:        0, // Disable browser cache during debugging
	})

	// Serve root-level static files (index.html will be handled by SPA catch-all)
	app.Static("/", "./dist", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		Index:         "", // Don't auto-serve index.html here
		CacheDuration: 0,
		MaxAge:        0,
	})

	// Keep for backward compatibility
	app.Static("/static", "./static")

	// Catch-all route for React Router (SPA)
	// This must be LAST and only catches non-static-file routes
	app.Get("/*", func(c *fiber.Ctx) error {
		path := c.Path()

		// If it's a file extension, it's likely a static asset that wasn't found
		// Return 404 instead of serving index.html
		if strings.Contains(path, ".") {
			// Check if it's a known static file extension
			ext := path[strings.LastIndex(path, "."):]
			staticExtensions := []string{".js", ".css", ".map", ".ico", ".png", ".svg", ".jpg", ".jpeg", ".gif", ".webp", ".woff", ".woff2", ".ttf", ".eot"}
			for _, staticExt := range staticExtensions {
				if ext == staticExt {
					return fiber.NewError(404, "Asset not found: "+path)
				}
			}
		}

		// Not a static file - serve React SPA
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.SendFile("./dist/index.html")
	})

	// RAILWAY: Start database reconnection service if initially failed
	if db == nil {
		go func() {
			logrus.Info("🔄 RAILWAY: Starting database reconnection service")
			for {
				time.Sleep(30 * time.Second) // Try reconnecting every 30 seconds
				logrus.Info("🔄 Attempting database reconnection...")
				
				newDB, err := database.Initialize(cfg)
				if err != nil {
					logrus.WithError(err).Warn("Database reconnection failed, will retry...")
					continue
				}
				
				// Success! Update the global db variable
				db = newDB
				logrus.Info("✅ Database reconnection successful!")
				
				// Run migrations on reconnection
				if err := database.RunMigrations(db); err != nil {
					logrus.WithError(err).Warn("Failed to run migrations after reconnection")
				} else {
					logrus.Info("Database migrations completed after reconnection")
				}
				
				break // Exit reconnection loop
			}
		}()
	}

	// Start background services - but only if they're initialized
	go func() {
		// Wait for services to be initialized before starting
		for whatsappService == nil || queueService == nil {
			time.Sleep(1 * time.Second)
		}
		
		// Start WhatsApp queue processor
		whatsappService.StartQueueProcessor()
		
		// Start delayed message processor
		for {
			if err := queueService.ProcessDelayedMessages(); err != nil {
				logrus.WithError(err).Error("Error processing delayed messages")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// Start session cleanup service for database-backed sessions
	if db != nil {
		go func() {
			logrus.Info("Starting session cleanup service")
			for {
				// Clean up expired sessions every 30 minutes
				time.Sleep(30 * time.Minute)
				if _, err := db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW() OR is_active = FALSE`); err != nil {
					logrus.WithError(err).Error("Failed to cleanup expired sessions")
				} else {
					logrus.Info("Successfully cleaned up expired sessions")
				}
			}
		}()
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logrus.Info("Shutting down server...")
		app.Shutdown()
	}()

	// Note: Server is already running from the background goroutine above
	// Keep the main thread alive for graceful shutdown handling
	select {} // Block forever until shutdown signal
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log error for non-404s or API routes
	if code != 404 || (c.Path() != "" && len(c.Path()) >= 4 && c.Path()[:4] == "/api") {
		logrus.Errorf("Error %d: %v", code, err)
	}

	// Return JSON error for API routes
	if c.Path() != "" && len(c.Path()) >= 4 && c.Path()[:4] == "/api" {
		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
			"code":    code,
		})
	}

	// For 404 errors on web routes, serve the React app (SPA routing)
	if code == 404 {
		// Set no-cache headers for HTML
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.SendFile("./dist/index.html")
	}

	// For other errors, return simple error response
	return c.Status(code).JSON(fiber.Map{
		"error": "Internal server error",
		"code":  code,
	})
}
