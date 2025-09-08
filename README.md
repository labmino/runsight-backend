# RunSight Backend 🏃‍♂️

**RunSight Backend** is a REST API service built with Go that manages user data, device pairing, and run tracking for smart running devices. It provides secure authentication, device management, and data storage capabilities for IoT devices and mobile applications.

![Go](https://img.shields.io/badge/Go-1.25+-blue)
![Gin](https://img.shields.io/badge/Gin-v1.10-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-blue)
![License](https://img.shields.io/badge/License-MIT-yellow)

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [Deployment](#deployment)

## Features

* **User Management** – Register, log in, and manage user profiles securely.
* **Device Pairing** – Pair mobile apps with IoT running devices seamlessly.
* **Run Tracking** – Upload, store, and retrieve detailed running session data.
* **Statistics API** – Get personalized running stats and performance insights easily.


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

**Environment Variables**:
```env
# Server
PORT=8080
GIN_MODE=debug

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=runsight_user
DB_PASSWORD=runsight_password
DB_NAME=runsight
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key
```

## API Endpoints

**Base URL:** `http://localhost:8080/api/v1`

**Authentication:**
- Mobile: `Authorization: Bearer <jwt_token>`
- IoT: `Authorization: <device_token>`

**Auth APIs:**
- `POST /auth/register` - Register user
- `POST /auth/login` - User login
- `GET /auth/profile` - Get profile (JWT required)

**Mobile APIs** (JWT required):
- `POST /mobile/pairing/request` - Request pairing code
- `GET /mobile/pairing/{id}/status` - Check pairing status
- `GET /mobile/devices` - List paired devices
- `GET /mobile/runs` - List runs (paginated)
- `GET /mobile/runs/{id}` - Get run details
- `GET /mobile/stats` - Get statistics

**IoT APIs** (device token required):
- `POST /iot/pairing/verify` - Verify pairing code
- `POST /iot/runs/upload` - Upload run data
- `POST /iot/runs/batch` - Batch upload runs
- `POST /iot/devices/status` - Update device status

**Monitoring:**
- `GET /health` - Health check


## Development

**Project Structure:**
```
cmd/server/main.go          # Entry point
internal/
├── handlers/               # HTTP handlers  
├── models/                 # Database models
├── services/               # Business logic
└── utils/                  # JWT, logging, responses
```

**Run locally:**
```bash
go run cmd/server/main.go
```

## Deployment

**Docker (recommended):**
```bash
docker-compose up -d
```

**Manual deployment:**
```bash
go build -o server cmd/server/main.go
GIN_MODE=release ./server
```

## License

MIT License - see [LICENSE](LICENSE) file.