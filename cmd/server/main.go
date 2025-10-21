package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
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
			"status": "ok",
			"message": "Chatbot Automation Platform - Rebuilt from scratch",
		})
	})

	// API routes
	api := app.Group("/api")
	api.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "running",
			"version": "2.0.0-rebuild",
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
