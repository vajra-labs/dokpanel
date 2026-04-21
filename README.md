<div align="center">
  <h1>🚀 dokpanel</h1>
  <p><strong>A lightweight, high-performance deployment platform built with Go, Fiber v3, and React + TanStack Router</strong></p>
  <p>Self-hostable Platform as a Service (PaaS) for modern application deployment</p>
</div>

<br />

dokpanel is a free, self-hostable deployment platform that simplifies application and database management with blazing-fast performance powered by Go.

## ✨ Features

- **Lightning Fast**: Built with Go and Fiber v3 for maximum performance
- **Docker Native**: Deploy and manage Docker containers with ease
- **Database Support**: Built-in support for PostgreSQL, MySQL, MongoDB, Redis, and SQLite
- **RESTful API**: Complete API for automation and integrations
- **Structured Logging**: Production-ready logging with Zerolog
- **Security First**: Helmet middleware, CORS, rate limiting, and secure defaults
- **Environment Management**: Validated config with go-playground/validator
- **Built-in Dashboard**: React (TanStack Router) frontend embedded in Go binary — single binary deploy
- **Minimal Footprint**: ~13MB binary, <50MB idle memory
- **Auto Recovery**: Built-in panic recovery for production stability
- **Health Monitoring**: Real-time health checks with memory stats

## 🚀 Getting Started

### Prerequisites

- Go 1.23+
- [Bun](https://bun.com/) (for building the frontend)
- [Biome](https://biomejs.dev/guides/manual-installation/) (for linting + formatting)
- [Taskfile](https://taskfile.dev/docs/installation) (cross-platform build tool)
- Docker (optional)

### Installation

```bash
git clone https://github.com/vajra-labs/dokpanel.git
cd dokpanel
task web:deps
```

### Development

```bash
task dev
```

Server starts at `http://localhost:8000`.

### Production Build

```bash
task build   # builds React SPA + embeds into Go binary
task start   # runs the binary
```

## 🛠️ Available Commands

```bash
task              # Show all available commands
task dev          # Start dev server with live reload (Air)
task build        # Build production binary (includes web:build)
task start        # Run production binary
task code:test    # Run all tests
task code:format  # Format Go source code
task mod:deps     # Download Go dependencies
task mod:tidy     # Tidy go.mod
task mod:clean    # Remove build artifacts

# Web dashboard
task web:dev      # Start React dev server (port 3000)
task web:build    # Build React SPA for production
task web:deps     # Install frontend dependencies
task web:lint     # Lint with Biome
task web:format   # Format with Biome
task web:check    # Biome check (lint + format)

# Database migrations (goose)
task migrate:up      # Run pending migrations
task migrate:down    # Rollback last migration
task migrate:status  # Show migration status
task migrate:reset   # Rollback all migrations

# Code generation (sqlc)
task sqlc         # Generate type-safe Go from SQL
```

## 🔧 Configuration

Configure via `.env` file:

```env
GO_ENV="development"         # development | production
HOST="0.0.0.0"
PORT=8000
SECRET="your-secret-key-min-32-chars"
CORS_ALLOW_ORIGIN="http://localhost:3000"
BODY_LIMIT="2MB"
DB_PATH="./dokpanel.db"

# JWT
JWT_ACCESS_EXP="5m"
JWT_REFRESH_EXP="24h"

# Rate limiting
RATE_LIMIT_MAX_REQ=100
RATE_LIMIT_WINDOWS="15m"

# Docker
DOCKER_HOST="unix:///var/run/docker.sock"
DOCKER_API_VERSION="1.41"
```

## 🏗️ Architecture

**Handler → Service → Repository → Database**

```
src/
├── apis/          # Route handlers (health, ...)
├── conf/          # Config loading & validation
├── consts/        # Shared constants & enums
├── db/            # Database connection
├── lib/           # Shared utilities (HttpError, ...)
├── logger/        # Zerolog setup
├── middle/        # Middleware (error, rate limit, 404)
├── server/        # Fiber app setup
└── main.go

webui/             # React dashboard (TanStack Router + Tailwind)
├── src/
│   ├── routes/    # File-based routes
│   └── main.tsx
└── embed.go       # Embeds dist/ into Go binary

tests/             # Integration tests
db/                # SQL migrations (goose) & sqlc config
```

### `webui/` — Frontend Dashboard

- **TanStack Router** — file-based routing, SPA mode
- **React Compiler** — automatic memoization
- **Tailwind CSS v4** — utility-first styling
- **Biome** — linting + formatting
- **Embedded in Go binary** via `//go:embed` — single binary deploy
- **Routing**: `/api/*` handled by Go, everything else served by React SPA

## 🔐 Security

- Helmet middleware for security headers
- CORS with credential support
- Request body size limits
- Rate limiting per IP
- Panic recovery middleware
- Config validation on startup

## 📝 API

```bash
GET /api/ping    → "Pong!"
GET /api/pong    → "Ping!"
GET /api/health  → { uptime, version, environment, timestamp, memory }
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

## 👨‍💻 Author

**Aashish Panchal** · [GitHub @vajra-labs](https://github.com/vajra-labs) · aipanchal51@gmail.com

---

<div align="center">
  <p>Made with ❤️ using Go</p>
  <p>⭐ Star this repo if you find it useful!</p>
</div>
