# 🌐 Real-Time Forum

A modern forum application with real-time private messaging capabilities, built with Go and vanilla JavaScript.

---

## Real-Time Forum Status: Complete (v1.0)

**Progress:** Backend Complete ✅ | Frontend Complete ✅ | Real-Time Messaging Active 🚀

---

## 🎯 Project Overview

This is a modern Single Page Application (SPA) forum featuring:
- **Real-time private messaging** using WebSockets
- **Instant updates** for online/offline user status
- **Dynamic content loading** (no page refreshes)
- **RESTful JSON API** backend
- **Premium UI** with responsive design

---

## 👥 Team

- **Naveed Minhas (mminhas)** - Team Lead, Full Stack Developer
- **Musab Shoaib (mshoaib)** - Frontend Developer

---

## 🏗️ Project Structure

```
real-time-forum/
├── backend/                          # Go backend server
│   ├── cmd/server/main.go           # Server entry point
│   ├── internal/
│   │   ├── database/                # SQLite DB & Models
│   │   ├── handlers/                # JSON API Handlers
│   │   ├── middleware/              # Auth Middleware
│   │   └── websocket/               # Real-time Hub
│   ├── forum.db                     # SQLite database
│   └── go.mod                       # Dependencies
│
├── frontend/                         # Vanilla JS SPA
│   ├── index.html                   # App Shell
│   ├── css/styles.css               # Premium Styling
│   └── js/
│       ├── app.js                   # Router & State
│       ├── api.js                   # API Client
│       ├── chat.js                  # Real-time Logic
│       ├── views.js                 # HTML Templates
│       └── websocket.js             # WS Connection
│
└── README.md                        # Documentation
```

---

## ✅ Completed Features

### Backend (Go)
- **JSON API Architecture**: Fully refactored from HTML templates to RESTful JSON endpoints.
- **Authentication**: Secure registration/login with bcrypt and session cookies.
- **WebSocket Hub**: Robust connection management with concurrent client handling.
- **Database**: SQLite with optimized schema for users, posts, comments, and messages.

### Frontend (JavaScript SPA)
- **Single Page Application**: Hash-based routing for instant navigation.
- **Real-Time Chat**: Private messaging with live history and scroll throttling.
- **User Tracking**: Live "Online" status indicators.
- **Modern UI**: Clean, responsive design with floating chat windows.

---

## 🚀 Quick Start

```bash
# Clone
git clone https://learn.01founders.co/git/mminhas/real-time-forum.git
cd real-time-forum

# Install dependencies & Run
cd backend
go mod download
go run cmd/server/main.go

# Access
# App: http://localhost:8080
```

---

## 📡 API Endpoints

```
POST   /register              - Create account
POST   /login                 - Login
GET    /posts                 - List posts (JSON)
POST   /posts/create          - Create post
WS     /ws                    - WebSocket Stream
POST   /api/messages/send     - Send DM
GET    /api/messages/history  - Get Chat History
GET    /api/online-users      - Get Online List
```

---

**Status:** Production Ready 🚀  
**Version:** 1.0.0  
**Last Updated:** November 26, 2025
