package server

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"pipeclaw/core/auth"
	"pipeclaw/core/pipeline"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var pipelineMgr *pipeline.Manager

// Setup error logging
func init() {
	// Create logs directory
	os.MkdirAll("logs", 0755)

	// Setup file logging
	file, err := os.OpenFile("logs/pipeclaw.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("❌ Failed to open log file: %v", err)
		return
	}

	// Multi-output: console + file
	multiWriter := &multiWriter{writers: []interface{ Write([]byte) (int, error) }{os.Stdout, file}}
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type multiWriter struct {
	writers []interface{ Write([]byte) (int, error) }
}

func (m *multiWriter) Write(p []byte) (int, error) {
	for _, w := range m.writers {
		w.Write(p)
	}
	return len(p), nil
}

// Start launches the web server
func Start(port string) error {
	pipelineMgr = pipeline.NewManager()

	log.Printf("🌊 PipeClaw server starting on port %s", port)
	log.Printf("📡 Dashboard: http://0.0.0.0:%s", port)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("❌ HTTP Error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
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
	app.Use("/ws/chat", websocket.New(handleChatWS, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))

	// Static files (dashboard)
	app.Static("/", "public")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
	})

	// Catch-all for 404
	app.Use(func(c *fiber.Ctx) error {
		log.Printf("⚠️  404: %s %s", c.Method(), c.Path())
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	})

	err := app.Listen(":" + port)
	if err != nil {
		log.Printf("❌ Server failed to start: %v", err)
		return err
	}

	log.Printf("✅ Server running successfully")
	return nil
}

func handleLogin(c *fiber.Ctx) error {
	log.Printf("🔐 Login attempt from %s", c.IP())

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&creds); err != nil {
		log.Printf("❌ Login parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	admin, err := auth.ValidateAdmin(creds.Username, creds.Password)
	if err != nil {
		log.Printf("❌ Login failed for %s: %v", creds.Username, err)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check if password needs to be changed
	if !admin.PasswordChanged {
		log.Printf("ℹ️  Password change required for %s", creds.Username)
		return c.JSON(fiber.Map{
			"success":                  false,
			"password_change_required": true,
		})
	}

	log.Printf("✅ Login successful for %s", creds.Username)
	return c.JSON(fiber.Map{
		"success":  true,
		"username": admin.Username,
	})
}

func handleChangePassword(c *fiber.Ctx) error {
	log.Printf("🔐 Password change request")

	var req struct {
		NewPassword string `json:"new_password"`
		Confirm     string `json:"confirm"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Password change parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.NewPassword != req.Confirm {
		log.Printf("❌ Passwords do not match")
		return c.Status(400).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	if len(req.NewPassword) < 4 {
		log.Printf("❌ Password too short")
		return c.Status(400).JSON(fiber.Map{"error": "Password too short"})
	}

	if err := auth.UpdatePassword(req.NewPassword); err != nil {
		log.Printf("❌ Password update failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	log.Printf("✅ Password changed successfully")
	return c.JSON(fiber.Map{"success": true})
}

func handleAuthStatus(c *fiber.Ctx) error {
	admin, err := auth.GetAdminStatus()
	if err != nil {
		log.Printf("❌ Auth status error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get status"})
	}

	return c.JSON(fiber.Map{
		"password_change_required": !admin.PasswordChanged,
	})
}

func handleGetPipeline(c *fiber.Ctx) error {
	log.Printf("📊 Getting pipeline")

	pipeline, err := pipelineMgr.Load()
	if err != nil {
		log.Printf("❌ Pipeline load error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load pipeline"})
	}

	return c.JSON(pipeline)
}

func handleSavePipeline(c *fiber.Ctx) error {
	log.Printf("💾 Saving pipeline")

	var pipeline pipeline.Pipeline
	if err := c.BodyParser(&pipeline); err != nil {
		log.Printf("❌ Pipeline save parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Save(&pipeline); err != nil {
		log.Printf("❌ Pipeline save error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save pipeline"})
	}

	log.Printf("✅ Pipeline saved")
	return c.JSON(fiber.Map{"success": true})
}

func handleDeploy(c *fiber.Ctx) error {
	log.Printf("🚀 Deploying pipeline")

	var req struct {
		PipelineID string `json:"pipeline_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Deploy parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Deploy(req.PipelineID); err != nil {
		log.Printf("❌ Deploy error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("✅ Pipeline deployed: %s", req.PipelineID)

	// Broadcast deployment event to all connected clients
	broadcastDeployment(true)

	return c.JSON(fiber.Map{"success": true})
}

func handleUndeploy(c *fiber.Ctx) error {
	log.Printf("⬛ Undeploying pipeline")

	var req struct {
		PipelineID string `json:"pipeline_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Undeploy parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := pipelineMgr.Undeploy(req.PipelineID); err != nil {
		log.Printf("❌ Undeploy error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("✅ Pipeline undeployed: %s", req.PipelineID)

	broadcastDeployment(false)

	return c.JSON(fiber.Map{"success": true})
}

// WebSocket handlers
var connectedClients = make(map[*websocket.Conn]bool)
var clientsMu = &sync.Mutex{}

func handleChatWS(c *websocket.Conn) {
	log.Printf("🔌 WebSocket connected from %s", c.RemoteAddr())

	clientsMu.Lock()
	connectedClients[c] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(connectedClients, c)
		clientsMu.Unlock()
		c.Close()
		log.Printf("🔌 WebSocket disconnected")
	}()

	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("⚠️  WebSocket error: %v", err)
			break
		}

		if msgType == websocket.TextMessage {
			log.Printf("💬 Received: %s", string(message))
			handleChatMessage(c, string(message))
		}
	}
}

func handleChatMessage(conn *websocket.Conn, message string) {
	// Process message through pipeline
	response := processPipeline(message)
	log.Printf("🤖 Response: %s", response)

	// Send response back
	conn.WriteJSON(fiber.Map{
		"type":    "response",
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
