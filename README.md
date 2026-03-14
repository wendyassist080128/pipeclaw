# 🌊 PipeClaw

A visual pipeline orchestration platform with drag-and-drop canvas editor.

## 🚀 Quick Start

### 1. Install & Run

```bash
# Clone the repository
git clone https://github.com/EasonChow/pipeclaw.git
cd pipeclaw

# Download dependencies
go mod download

# Run PipeClaw
go run cmd/main.go
```

### 2. First Run - Auto Setup

On first run, PipeClaw will:

```
🌊 PipeClaw - First Time Setup
===============================

Creating default admin account...
✅ Default admin created (username: admin, password: admin)

Creating default pipeline canvas...
✅ Default pipeline created

🌊 PipeClaw starting...
📡 Web Dashboard: http://0.0.0.0:23323

Access from your device:
  http://<your-ip>:23323

🔐 First login:
  Username: admin
  Password: admin

⚠️  You'll be forced to change password on first login!
```

### 3. Access Web Dashboard

Open your browser: `http://localhost:23323`

**Login:**
- Username: `admin`
- Password: `admin`

### 4. Change Password (Required)

After first login, you'll be prompted to change your password:

```
🔐 Change Your Password
- New Password: [your password]
- Confirm Password: [your password]
[Update Password]
```

### 5. Dashboard - Pipeline Canvas

You'll see the default pipeline:

```
┌─────────────────────────────────────────────────┐
│  Pipeline Canvas Editor                         │
│                                                 │
│  ┌────────────┐    ┌──────────────┐    ┌────────┐
│  │  [input]   │───▶│ [session:main]│───▶│[output]│
│  └────────────┘    └──────────────┘    └────────┘
│                                                 │
│  [🚀 Deploy]                                    │
└─────────────────────────────────────────────────┘
```

### 6. Deploy Pipeline

Click the **[Deploy]** button in the top right:

```
✅ Pipeline Deployed!
```

The status indicator turns **green** and a **floating chat widget** appears at the bottom right.

### 7. Chat with Wendy

The chat widget auto-sends:

```
🤖 "Wake up my friend!"
```

Now you can talk to Wendy Amira! 💬

---

## 📁 Project Structure

```
pipeclaw/
├── cmd/
│   └── main.go              # Entry point
│
├── core/
│   ├── auth/                # Authentication
│   │   └── auth.go
│   ├── pipeline/            # Pipeline management
│   │   └── manager.go
│   └── server/              # Web server
│       └── server.go
│
├── public/
│   └── index.html           # Web dashboard
│
├── data/
│   ├── admin.json           # Admin credentials
│   └── pipeline.json        # Pipeline config
│
├── go.mod
└── README.md
```

---

## 🎯 Features

✅ **Visual Pipeline Editor** - Drag-and-drop canvas  
✅ **Real-time Deployment** - Instant deploy/undeploy  
✅ **Floating Chat Widget** - Live chat when deployed  
✅ **Secure Auth** - Password change on first login  
✅ **WebSocket** - Real-time communication  
✅ **Single Binary** - No dependencies needed  

---

## 🔧 Configuration

### Environment Variables

```bash
export PIPECLAW_PORT=23323    # Default port
export PIPECLAW_HOST=0.0.0.0  # Default host
```

### Custom Pipeline

Edit `data/pipeline.json` to customize nodes and edges:

```json
{
  "id": "main",
  "name": "Main Pipeline",
  "nodes": [
    { "id": "input", "type": "input", "label": "Input", "x": 50, "y": 100 },
    { "id": "session:main", "type": "session", "label": "Session: Main", "x": 250, "y": 100 },
    { "id": "output", "type": "output", "label": "Output", "x": 450, "y": 100 }
  ],
  "edges": [
    { "from": "input", "to": "session:main" },
    { "from": "session:main", "to": "output" }
  ],
  "deployed": false
}
```

---

## 📡 API Reference

### Authentication

```http
POST /api/login
{
  "username": "admin",
  "password": "admin"
}
```

```http
POST /api/change-password
{
  "new_password": "newpass",
  "confirm": "newpass"
}
```

### Pipeline

```http
GET /api/pipeline          # Get current pipeline
POST /api/pipeline         # Save pipeline
POST /api/pipeline/deploy  # Deploy pipeline
POST /api/pipeline/undeploy # Undeploy pipeline
```

### WebSocket

```
ws://localhost:23323/ws/chat
```

Send message:
```json
{"type": "message", "content": "Hello"}
```

Receive response:
```json
{"type": "response", "message": "🤖 Wake up my friend!"}
```

---

## 🚀 Deploy to RTX Server

```bash
# 1. SSH to RTX
ssh sshuser@192.168.100.62

# 2. Clone PipeClaw
git clone https://github.com/EasonChow/pipeclaw.git
cd pipeclaw

# 3. Install Go dependencies
go mod download

# 4. Run
go run cmd/main.go

# 5. Access from your PC
# Open: http://192.168.100.62:23323
```

---

## 🎨 Pipeline Node Types

| Type | Description |
|------|-------------|
| `input` | Message input (chat, API, etc.) |
| `session` | Session management |
| `output` | Response output |
| `llm` | LLM processing |
| `tool` | Tool execution |
| `condition` | Conditional branching |

---

## 📝 Next Steps

1. **Add more node types** (LLM, tools, conditions)
2. **Save multiple pipelines**
3. **Pipeline versioning**
4. **Team collaboration**
5. **Plugin system**

---

## License

MIT
