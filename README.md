<div align="center">
  <h1>🚀 dokpanel</h1>
  <p><strong>A lightweight, high-performance deployment platform built with Go, Fiber v3, and React + TanStack Router</strong></p>
  <p>Self-hostable Platform as a Service (PaaS) for modern application deployment</p>
</div>

> [!WARNING]
> This project is currently in **active development** and is not ready for production use. APIs and configurations are subject to change.

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
- goqite->(JobQueue), gopsutil->(Monitoring), and robfig/cron->(Backups/Cleanup)

## 🚀 Getting Started

### Prerequisites

- Go 1.26+
- [Docker](https://www.docker.com/products/docker-desktop)
- [Bun](https://bun.com/) (for building the frontend)
- [Taskfile](https://taskfile.dev/docs/installation) (cross-platform build tool)
- [Atlas](https://atlasgo.io/getting-started#step-1-install-atlas) (optional, for schema diffing)

### Installation

```bash
git clone https://github.com/vajra-labs/dokpanel.git
cd dokpanel
task web:deps
```

### Development

1. **One-time Setup**: Run the setup command to initialize the development environment (Docker Swarm, Traefik, etc.):

   ```bash
   task setup
   ```

2. **Start Dev Server**: Start the development server with live reload (Air):

   ```bash
   task dev
   ```

   Server starts at `http://localhost:8000`.

3. **Teardown**: Revert the dev setup (remove Traefik, Swarm, etc.):
   ```bash
   task teardown
   ```

### Production Build

```bash
task build   # builds React SPA + embeds into Go binary
task start   # runs the binary
```

## 🛠️ Available Commands

```bash
task              # Show all available commands
task setup        # One-time dev setup (Swarm, Traefik, etc.)
task teardown     # Revert dev setup (remove Traefik, etc.)
task dev          # Start dev server with live reload (Air)
task build          # Build production binary (includes web:build)
task server:build   # Build Go server binary only (skips web:build)
task start          # Run production binary
task code:test    # Run all tests
task code:format  # Format Go source code
task mod:deps     # Download Go dependencies
task mod:tidy     # Tidy go.mod
task mod:clean    # Remove build artifacts

# Web dashboard
task web:dev      # Start React dev server (port 3000)
task web:build    # Build React SPA for production
task web:deps     # Install frontend dependencies
task web:lint     # Lint with ESLint
task web:format   # Format with Prettier
task web:check    # Check (lint + format)

# Database migrations (goose)
task migrate:up      # Run pending migrations
task migrate:down    # Rollback last migration
task migrate:status  # Show migration status
task migrate:reset   # Rollback all migrations

# Schema management (atlas + goose)
task atlas:diff NAME=<label>  # Diff schema and generate Goose migration file

# Code generation (sqlc)
task sqlc         # Generate type-safe Go from SQL
```

## 🔧 Configuration

Configure via `.env` file:

```env
GO_ENV="development" # dev, prod, test
HOST="0.0.0.0"
PORT=8000
SECRET="your-secret-key-min-32-chars"
CORS_ALLOW_ORIGIN="http://localhost:3000"
BODY_LIMIT="2MB"
DB_PATH="sqldb/db.sqlite3"

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

**fx (dependency injection) → Handler → Repository → Database**

```
cmd/
├── server/        # Main server binary
│   └── main.go    # fx wiring: conf, logger, db, apis, fiber, web
└── setup/         # One-time setup CLI binary
    └── main.go    # fx wiring: conf, logger, docker, setup

src/
├── fiber.go       # Fiber app factory (middleware stack)
├── apis/          # Route registration & handlers
│   └── health/    # Health, ping, pong endpoints
├── conf/          # Config loading & validation (module.go + provider.go)
├── db/            # Database client & repositories (module.go + provide.go)
├── lib/
│   ├── docker/    # Docker client & paths (module.go + client.go)
│   ├── core/      # Shared error types
│   ├── process/   # Process utilities
│   └── utils/     # String helpers
├── logger/        # Zerolog setup (module.go + provide.go)
├── middle/        # Error handler, rate limiter, 404
├── setup/         # Docker Swarm, Traefik, network setup
└── types/         # Shared enums

web/               # React dashboard (TanStack Router + Tailwind v4)
├── src/
│   ├── routes/    # File-based routes
│   └── main.tsx
└── embed.go       # Embeds dist/ into Go binary

tests/             # Integration tests
sqldb/             # SQL schema, migrations & sqlc config
├── migrate/       # Goose migration files (embedded in binary)
├── schema/        # Atlas schema source files
├── queries/       # sqlc SQL queries
├── embed.go       # Embeds migrate/ + runs on startup via goose
└── tools/         # Atlas post-processor (trigger patch)
```

### fx Dependency Graph

```
conf.Module   →  logger.Module
              →  db.Module      →  sqldb.Migrate (on start)
              →  docker.Module  →  ping (on start), close (on stop)
              →  apis.Module    →  health handlers
              →  src.Fiber      →  middleware stack
              →  web.ServeSPA   →  embedded React SPA
              →  StartServer    →  listen (on start), shutdown (on stop)
```

### `web/` — Frontend Dashboard

- **TanStack Router** — file-based routing, SPA mode
- **React Compiler** — automatic memoization
- **Tailwind CSS v4** — utility-first styling
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
