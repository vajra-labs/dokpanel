# Tech Stack

Dokpanel is a lightweight, single-server, self-hosted deployment dashboard designed to monitor and manage Docker Swarm applications, relational databases, and compose stacks. The application is built with a Go backend (Fiber v3) and a modern React dashboard (React Router v7 + Bun).

---

## 1. Core Technologies

### 1.1 Frontend (Dashboard)

- **Framework**: React Router v7
- **Language**: TypeScript 5.x
- **Package Manager**: Bun
- **Styling**: Tailwind CSS
- **Build Tool**: Vite / React Router compiler

### 1.2 Backend (Server)

- **Framework**: Fiber v3 (`github.com/gofiber/fiber/v3`)
- **Language**: Go (Golang) 1.26.3
- **Dependency Injection**: Uber Fx (`go.uber.org/fx`)
- **JSON & Utilities**: `dustin/go-humanize`
- **Environment Loader**: `joho/godotenv`

### 1.3 API Documentation

- **Spec Generation**: Huma v2 (`github.com/danielgtaylor/huma/v2`) — generates OpenAPI 3.1 spec from Go structs
- **UI Renderer**: go-scalar-api-reference (`github.com/MarceloPetrucio/go-scalar-api-reference`) — Scalar UI at `/api/docs`
- **Code Quality**: golangci-lint v2 + gofumpt for linting and formatting

### 1.3 Database & Query Layer

- **Database**: SQLite 3 (STRICT tables enabled)
- **Database Driver**: `mattn/go-sqlite3`
- **Query Compiler**: SQLC (generates type-safe Go repository code from raw SQL queries)
- **Migration Engine**: Goose (`github.com/pressly/goose/v3`)
- **Schema Diffing**: Atlas CLI (generates migrations by comparing schema DDL files)

---

## 2. Infrastructure & Orchestration

### 2.1 Docker & Swarm Control

- **Libraries**:
  - Moby API (`github.com/moby/moby/api`)
  - Moby Client (`github.com/moby/moby/client`)
- **Mechanics**:
  - Direct communication with the host Unix socket (`/var/run/docker.sock`)
  - Orchestrating Swarm services (replicated modes, network configs, replica counts)
  - Dynamic volume mounts (binds, persistent volumes)

### 2.2 Network & Routing

- **Reverse Proxy**: Traefik (Docker Swarm provider integration)
- **SSL Auto-Provisioning**: Let's Encrypt / ACME resolvers via Traefik configurations

---

## 3. Security & Validation

### 3.1 Authentication & Authorization

- **Token Strategy**: Stateless JSON Web Tokens (JWT)
- **Cryptography**: `golang.org/x/crypto` (Bcrypt for password hashing)
- **Session Control**: Token blacklist tracking via database-stored JTI records

### 3.2 Request Validation

- **Engine**: Go Playground Validator v10 (`github.com/go-playground/validator/v10`)

---

## 4. Development & Logging

### 4.1 Development Workflow

- **Task Runner**: Taskfile (`Taskfile.yml`) to orchestrate migration updates, code gen, and builds
- **Hot Reloading**: Air (`air-verse/air`) for rapid Go backend live-reloads

### 4.2 Logging & Rolling

- **Structured Logging**: Zerolog (`github.com/rs/zerolog`)
- **Log Rotation**: Lumberjack (`gopkg.in/natefinch/lumberjack.v2`)

### 4.3 Testing Stack

- **Framework**: `stretchr/testify` for assertions and mockings
- **Tooling**: `gotestsum` for formatted test outputs
