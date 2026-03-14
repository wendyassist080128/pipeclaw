package main

import (
	"fmt"
	"log"
	"os"

	"pipeclaw/core/auth"
	"pipeclaw/core/pipeline"
	"pipeclaw/core/server"
	"pipeclaw/internal/logger"
)

func main() {
	logger.Init()

	// Check if first run (no admin user)
	isFirstRun := !auth.AdminExists()

	if isFirstRun {
		fmt.Println("🌊 PipeClaw - First Time Setup")
		fmt.Println("===============================")
		fmt.Println()
		fmt.Println("Creating default admin account...")
		
		// Create default admin
		if err := auth.CreateAdmin("admin", "admin"); err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
		
		fmt.Println("✅ Default admin created (username: admin, password: admin)")
		fmt.Println()
	}

	// Load or create default pipeline
	pipelineMgr := pipeline.NewManager()
	if isFirstRun {
		fmt.Println("Creating default pipeline canvas...")
		
		defaultPipeline := &pipeline.Pipeline{
			ID:   "main",
			Name: "Main Pipeline",
			Nodes: []pipeline.Node{
				{ID: "input", Type: "input", Label: "Input"},
				{ID: "session:main", Type: "session", Label: "Session: Main"},
				{ID: "output", Type: "output", Label: "Output"},
			},
			Edges: []pipeline.Edge{
				{From: "input", To: "session:main"},
				{From: "session:main", To: "output"},
			},
			Deployed: false,
		}
		
		if err := pipelineMgr.Save(defaultPipeline); err != nil {
			log.Fatalf("Failed to save default pipeline: %v", err)
		}
		
		fmt.Println("✅ Default pipeline created")
		fmt.Println()
	}

	// Start server
	port := os.Getenv("PIPECLAW_PORT")
	if port == "" {
		port = "23323"
	}

	fmt.Println("🌊 PipeClaw starting...")
	fmt.Printf("📡 Web Dashboard: http://0.0.0.0:%s\n", port)
	fmt.Println()
	fmt.Println("Access from your device:")
	fmt.Printf("  http://<your-ip>:%s\n", port)
	fmt.Println()

	if isFirstRun {
		fmt.Println("🔐 First login:")
		fmt.Println("  Username: admin")
		fmt.Println("  Password: admin")
		fmt.Println()
		fmt.Println("⚠️  You'll be forced to change password on first login!")
		fmt.Println()
	}

	if err := server.Start(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
