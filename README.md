# RunSight Backend

**RunSight Backend** is a REST API service for the RunSight smart running system. It manages user authentication, device pairing, and run data for IoT devices (smart glasses) and mobile applications.

![Go](https://img.shields.io/badge/Go-1.25+-blue)
![Gin](https://img.shields.io/badge/Gin-v1.10-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-blue)
![License](https://img.shields.io/badge/License-MIT-yellow)

## System Overview

RunSight is an autonomous IoT-first running assistance system where smart glasses provide real-time AI guidance to runners. The backend serves as the central data hub that:

- Manages secure device pairing between mobile apps and IoT devices
- Stores running sessions with AI metrics (obstacle detection, lane keeping, warnings)
- Provides offline-first sync capabilities for IoT devices
- Delivers personalized statistics and run history to mobile users

**Core Principles:**
- IoT devices are autonomous (runs happen without mobile connectivity)
- Mobile apps are view-only (read history, manage devices)
- Backend is stateless with optional real-time features
- Offline-first with automatic sync when network is available

## Features

* **Secure Authentication** – JWT-based user auth and device token management
* **Device Pairing** – 6-digit code pairing system for mobile-IoT connection
* **Run Data Management** – Upload, store, and sync running sessions with AI metrics
* **Statistics & Analytics** – Aggregated performance insights and history
* **Monitoring & Health** – Health checks and structured logging


## Quick Start

```bash
git clone https://github.com/labmino/runsight-backend.git
cd runsight-backend

# Using Docker (recommended)
docker-compose up -d

# Or manual setup
cp .env.example .env
go run cmd/server/main.go
```

Server runs on `http://localhost:8080` with PostgreSQL database.

## API Endpoints

**Base URL:** `http://localhost:8080/api/v1`

### Authentication
- Mobile apps: `Authorization: Bearer <jwt_token>`
- IoT devices: `Authorization: Bearer <device_token>`

### Key Endpoints
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /mobile/pairing/request` - Mobile requests pairing code
- `POST /iot/pairing/verify` - IoT verifies pairing code
- `POST /iot/runs/upload` - Upload run data with AI metrics
- `GET /mobile/runs` - Get user's run history
- `GET /mobile/stats` - Get aggregated statistics
- `GET /health` - Health check


## Development

**Project Structure:**
```
cmd/server/main.go          # Entry point
internal/
├── handlers/               # HTTP handlers (auth, mobile, iot, monitoring)
├── models/                 # Database models (user, device, run, ai_metrics)
├── services/               # Business logic (pairing)
├── middleware/             # Auth, rate limiting, security
├── database/               # PostgreSQL connection and migrations
└── utils/                  # JWT, logging, responses, error codes
tests/                      # Unit and integration tests
```

**Run locally:**
```bash
go run cmd/server/main.go
```

**Run tests:**
```bash
go test ./...
```

## Deployment

**Docker (recommended):**
```bash
docker-compose up -d
```

**Production:**
```bash
go build -o server cmd/server/main.go
GIN_MODE=release ./server
```

## License

MIT License - see [LICENSE](LICENSE) file.