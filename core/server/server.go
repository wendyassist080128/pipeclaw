package server

import (
	"encoding/json"
	"log"
	"net/http"

	"pipeclaw/core/auth"
	"pipeclaw/core/pipeline"
	"pipeclaw/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var pipelineMgr *pipeline.Manager

// Start launches the web server
func Start(port string) error {
	pipelineMgr = pipeline.NewManager()

	app := fiber.New(fiber.Config{
		Views: nil, // Static files served directly
	})

	// API Routes
	api := app.Group("/api")

	// Auth routes
	api.Post("/login", handleLogin)
	api.Post("/change-password", handleChangePassword)
	api.Get("/auth-status", handleAuthStatus)

	// Pipeline routes
	api.Get("/pipeline", handleGetPipeline)
	api.Post("/pipeline", handleSavePipeline)
	api.Post("/pipeline/deploy", handleDeploy)
	api.Post("/pipeline/undeploy", handleUndeploy)

	// WebSocket for real-time chat
	app.Use("/ws/chat", websocket.New(handleChatWS))

	// Static files (dashboard)
	app.Static("/", "public")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Printf("🌊 PipeClaw server starting on port %s", port)
	return app.Listen(":" + port)
}

func handleLogin(c *fiber.Ctx) error {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&creds); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	admin, err := auth.ValidateAdmin(creds.Username, creds.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check if password needs to be changed
	if !admin.PasswordChanged {
		return c.JSON(fiber.Map{
			"success":          false,
			"password_change_required": true,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"username": admin.Username,
	})
}

func handleChangePassword(c *fiber.Ctx) error {
	var req struct {
		NewPassword string `json:"new_password"`
		Confirm     string `json:"confirm"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.NewPassword != req.Confirm {
		return c.Status(400).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	if len(req.NewPassword) < 4 {
		return c.Status(400).JSON(fiber.Map{"error": "Password too short"})
	}

	if err := auth.UpdatePassword(req.NewPassword); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func handleAuthStatus(c *fiber.Ctx) error {
	admin, err := auth.GetAdminStatus()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get status"})
	}

	return c.JSON(fiber.Map{
		"password_change_required": !admin.PasswordChanged,
	})
}

func handleGetPipeline(c *fiber.Ctx) error {
	pipeline, err := pipelineMgr.Load()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load pipeline"})
	}

	return c.JSON(pipeline)
}

func handleSavePipeline(c *fiber.Ctx) error {
	var pipeline pipeline.Pipeline
	if err := c.BodyParser(&pipeline); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Save(&pipeline); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save pipeline"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func handleDeploy(c *fiber.Ctx) error {
	var req struct {
		PipelineID string `json:"pipeline_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Deploy(req.PipelineID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Broadcast deployment event to all connected clients
	broadcastDeployment(true)

	return c.JSON(fiber.Map{"success": true})
}

func handleUndeploy(c *fiber.Ctx) error {
	var req struct {
		PipelineID string `json:"pipeline_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Undeploy(req.PipelineID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	broadcastDeployment(false)

	return c.JSON(fiber.Map{"success": true})
}

// WebSocket handlers
var connectedClients = make(map[*websocket.Conn]bool)
var clientsMu = &sync.Mutex{}

func handleChatWS(c *websocket.Conn) {
	clientsMu.Lock()
	connectedClients[c] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(connectedClients, c)
		clientsMu.Unlock()
		c.Close()
	}()

	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.TextMessage {
			// Handle chat message
			handleChatMessage(c, string(message))
		}
	}
}

func handleChatMessage(conn *websocket.Conn, message string) {
	// Process message through pipeline
	response := processPipeline(message)

	// Send response back
	conn.WriteJSON(fiber.Map{
		"type":   "response",
		"message": response,
	})
}

func processPipeline(message string) string {
	// Simple pipeline execution
	// input -> session:main -> output
	
	// For now, return a greeting
	return "🤖 Wake up my friend! How can I help you today?"
}

func broadcastDeployment(deployed bool) {
	message := fiber.Map{
		"type":     "deployment",
		"deployed": deployed,
	}

	data, _ := json.Marshal(message)

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range connectedClients {
		client.WriteMessage(websocket.TextMessage, data)
	}
}
