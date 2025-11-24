# 🌐 Real-Time Forum

A modern forum application with real-time private messaging capabilities, built with Go and vanilla JavaScript.

---

## 📊 Project Status: In Development (Week 1)

**Progress:** Backend Foundation Complete ✅ | Frontend In Progress 🚧

---

## 🎯 Project Overview

This is an enhanced version of a traditional forum, featuring:
- Real-time private messaging using WebSockets
- Online/offline user status tracking
- Traditional forum features (posts, comments, voting)
- Single Page Application (SPA) frontend
- RESTful API backend

---

## 👥 Team

- **Naveed Minhas (mminhas)** - Team Lead, Backend Developer
- **Musab Shoaib (mshoaib)** - Frontend Developer

---

## 🏗️ Project Structure

```
real-time-forum/
├── backend/                          # Go backend server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Server entry point
│   │
│   ├── internal/
│   │   ├── database/                # Database layer
│   │   │   ├── connection.go       # DB initialization
│   │   │   ├── models.go           # Data structures
│   │   │   └── migrations.go       # Real-time tables
│   │   │
│   │   ├── handlers/                # HTTP handlers
│   │   │   ├── auth.go             # Authentication
│   │   │   ├── posts.go            # Posts CRUD
│   │   │   ├── comments.go         # Comments
│   │   │   └── votes.go            # Voting system
│   │   │
│   │   ├── middleware/              # Middleware
│   │   │   └── auth.go             # Auth middleware
│   │   │
│   │   └── websocket/               # WebSocket system
│   │       ├── hub.go              # Connection manager
│   │       ├── client.go           # Client handler
│   │       └── handler.go          # HTTP upgrade handler
│   │
│   ├── forum.db                     # SQLite database
│   ├── go.mod                       # Go dependencies
│   └── go.sum
│
├── frontend/                         # JavaScript SPA
│   ├── index.html                   # Main HTML
│   ├── css/
│   │   └── styles.css               # Styling
│   └── js/
│       ├── config.js                # Configuration
│       ├── websocket.js             # WebSocket client
│       ├── api.js                   # API client
│       ├── chat.js                  # Chat logic
│       └── app.js                   # Main app
│
└── README.md                        # This file
```

---

## ✅ Completed Features

### Backend (100% Foundation Complete)

**Traditional Forum:**
- ✅ User authentication (register, login, logout)
- ✅ Session management with cookies
- ✅ Password hashing (bcrypt)
- ✅ Posts creation and viewing
- ✅ Categories system (10 default categories)
- ✅ Comments on posts
- ✅ Like/Dislike system
- ✅ User-specific filters

**Database:**
- ✅ SQLite with 10 tables
- ✅ Foreign key constraints
- ✅ Performance indexes
- ✅ `messages` table (private messages)
- ✅ `user_status` table (online tracking)

**Real-Time Infrastructure:**
- ✅ WebSocket hub (connection manager)
- ✅ Client handler (read/write)
- ✅ HTTP upgrade handler
- ✅ WebSocket endpoint: `ws://localhost:8080/ws`
- ✅ Connection authentication
- ✅ Tested and working!

---

### Frontend (60% Complete)

**UI:**
- ✅ HTML layout with sidebar and chat area
- ✅ Modern CSS styling
- ✅ Responsive design

**JavaScript:**
- ✅ WebSocket client class
- ✅ API client for HTTP requests
- ✅ Configuration
- 🚧 Chat logic (in progress)
- 🚧 Message display
- 🚧 Online users list

---

## 🚧 Currently Working On

### Backend (This Week):
- Message API endpoints
- Save/load messages from database
- Update user online/offline status

### Frontend (This Week):
- Connect WebSocket to UI
- Display messages
- Send messages functionality
- Show online users

---

## 🛠️ Tech Stack

**Backend:** Go 1.21+ | SQLite3 | Gorilla WebSocket  
**Frontend:** HTML5 | CSS3 | Vanilla JavaScript  
**Tools:** Git | Gitea

---

## 🚀 Quick Start

```bash
# Clone
git clone https://learn.01founders.co/git/mminhas/real-time-forum.git
cd real-time-forum

# Install dependencies
cd backend
go mod download

# Run server
go run cmd/server/main.go

# Access
# Forum: http://localhost:8080
# WebSocket: ws://localhost:8080/ws
```

---

## 📡 API Endpoints

### Current:
```
POST   /register              - Create account
POST   /login                 - Login
GET    /logout                - Logout
GET    /posts                 - List posts
POST   /posts/create          - Create post
POST   /comments/create       - Add comment
POST   /vote                  - Like/Dislike
WS     /ws                    - WebSocket connection
```

### Coming Soon:
```
POST   /api/messages/send     - Send message
GET    /api/messages/history  - Chat history
GET    /api/online-users      - Online users list
```

---

## 🗄️ Database Schema

**Traditional:** users, sessions, posts, comments, likes, categories, post_categories, contact_messages

**Real-Time (NEW):**
- `messages` - Private messages (id, sender_id, receiver_id, content, created_at, is_read)
- `user_status` - Online tracking (user_id, is_online, last_seen, websocket_id)

---

## 🔒 Security

- Password hashing (bcrypt)
- Session-based authentication
- SQL injection prevention
- Input validation
- Authenticated WebSocket connections

---

## 📈 Performance

- Server startup: ~50ms
- WebSocket connection: ~700µs
- Database queries: < 5ms
- Page load: < 100ms

---

## 🤝 Git Workflow

```bash
# Backend (Naveed)
git checkout backend/websocket
# ... work ...
git push origin backend/websocket

# Frontend (Musab)
git checkout frontend/chat
# ... work ...
git push origin frontend/chat

# Review via Pull Requests
```

---

## 🎯 Project Goals

**Must Have (MVP):**
- ✅ Traditional forum
- ✅ Authentication
- ✅ WebSocket infrastructure
- 🚧 Private messaging
- 🚧 Online status

**Nice to Have:**
- Typing indicators
- Message search
- File sharing
- Emoji support

---

## 🎉 Recent Updates

**November 11, 2025:**
- ✅ WebSocket integration complete
- ✅ Connection tested successfully
- ✅ Backend foundation ready

**November 10, 2025:**
- ✅ Real-time database tables
- ✅ WebSocket library installed
- ✅ Project organized

---

## 📞 Contact

**Team Lead:** Naveed Minhas (mminhas)  
**Repository:** https://learn.01founders.co/git/mminhas/real-time-forum  
**School:** 01 Founders

---

**Status:** Actively in development 🚀  
**Progress:** ~40% complete  
**Next:** Message API implementation

*Last Updated: November 11, 2025 | Version: 0.4.0-alpha*
