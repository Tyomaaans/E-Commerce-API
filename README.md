# E-Commerce API

A production-ready REST API for an e-commerce platform built with Go. Designed with scalability and maintainability in mind — featuring a multi-role system, event-driven architecture, and pluggable payment gateway abstraction.

> ⚠️ This project is actively under development. Core features are stable; payment and real-time chat are in progress.

---

## Architecture Overview

```
Client
  │
  ▼
Gin HTTP Server
  │
  ├── Auth Middleware (Redis — token validation, no DB round-trip)
  │
  ├── Handlers → Services → Repositories
  │                             │
  │                         PostgreSQL (primary datastore)
  │
  ├── Redis
  │     ├── Refresh token storage (hashed, with TTL)
  │     ├── OTP / verification token (register, forgot password)
  │     └── User cache
  │
  ├── Redpanda (Kafka-compatible event streaming)
  │     ├── Email notification worker & consumer
  │     ├── Audit log (persistent, replayable)
  │     └── Real-time chat broker (planned)
  │
  └── MinIO (object storage)
        ├── User avatars
        ├── Product images
        └── Store logos
```

---

## Tech Stack

| Layer | Technology | Reason |
|---|---|---|
| Language | Go | Performance, strong typing, excellent concurrency |
| Framework | Gin | Lightweight, fast HTTP router |
| Database | PostgreSQL | Reliable relational datastore |
| Cache & Token Store | Redis | Low-latency token validation; avoids DB round-trips on every auth check |
| Event Streaming | Redpanda | Kafka-compatible but lighter (no JVM); used for async notifications, audit log, and planned real-time chat |
| Object Storage | MinIO | S3-compatible self-hosted storage for media files |
| Containerization | Docker & Docker Compose | Consistent local and production environment |

---

## Features

### Auth
- Register with email verification
- Login with JWT (access token + refresh token)
- Refresh token rotation with SHA-256 hashing at rest
- Forgot password & reset password via email OTP
- Logout with token revocation

### Role-Based Access Control
Three roles with distinct permissions:

| Role | Capabilities |
|---|---|
| **Buyer** | Browse products, manage profile & addresses, register as seller |
| **Seller** | Manage own store and products (activate / deactivate) |
| **Admin** | Full management over users, stores, products, categories |

### Store & Product Management
- Sellers can register a store, update it, or deactivate it
- Products support soft delete (deactivate / reactivate) — hard delete is admin-only to preserve order history integrity
- Hierarchical categories with parent-child support

### Infrastructure Design Decisions
- **Redis** stores tokens as SHA-256 hashes — if Redis is compromised, raw tokens are not exposed
- **Redpanda** used over RabbitMQ for persistent audit log, message replay, and multiple independent consumer groups
- **Payment gateway** abstracted behind an interface — switching providers requires only a config change and API key swap, no logic changes

---

## API Endpoints

<details>
<summary><strong>Auth</strong> — 9 endpoints</summary>

| Method | Endpoint | Access |
|---|---|---|
| POST | `/api/v1/auth/register` | Public |
| POST | `/api/v1/auth/resend-verification` | Public |
| GET | `/api/v1/auth/verify-email` | Public |
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/auth/refresh` | Public |
| POST | `/api/v1/auth/forgot-password` | Public |
| GET | `/api/v1/auth/reset-password` | Public |
| POST | `/api/v1/auth/reset-password` | Public |
| POST | `/api/v1/auth/logout` | Authenticated |

</details>

<details>
<summary><strong>Users</strong> — 9 endpoints</summary>

| Method | Endpoint | Access |
|---|---|---|
| PUT | `/api/v1/users/me/profile` | Authenticated |
| DELETE | `/api/v1/users/me` | Authenticated |
| GET | `/api/v1/users/me` | Authenticated + Completed Profile |
| PATCH | `/api/v1/users/me` | Authenticated + Completed Profile |
| PUT | `/api/v1/users/me/password` | Authenticated + Completed Profile |
| GET | `/api/v1/users/me/addresses` | Authenticated + Completed Profile |
| POST | `/api/v1/users/me/addresses` | Authenticated + Completed Profile |
| PATCH | `/api/v1/users/me/addresses/:id` | Authenticated + Completed Profile |
| DELETE | `/api/v1/users/me/addresses/:id` | Authenticated + Completed Profile |

</details>

<details>
<summary><strong>Stores</strong> — 6 endpoints</summary>

| Method | Endpoint | Access |
|---|---|---|
| GET | `/api/v1/stores` | Public |
| GET | `/api/v1/stores/:id` | Public |
| POST | `/api/v1/stores` | Buyer |
| GET | `/api/v1/stores/me` | Seller |
| PATCH | `/api/v1/stores/me` | Seller |
| DELETE | `/api/v1/stores/me` | Seller |

</details>

<details>
<summary><strong>Products</strong> — 6 endpoints</summary>

| Method | Endpoint | Access |
|---|---|---|
| GET | `/api/v1/products` | Public |
| GET | `/api/v1/products/:id` | Public |
| POST | `/api/v1/stores/me/products` | Seller |
| PATCH | `/api/v1/stores/me/products/:id` | Seller |
| POST | `/api/v1/stores/me/products/:id/deactivate` | Seller |
| POST | `/api/v1/stores/me/products/:id/reactivate` | Seller |

</details>

<details>
<summary><strong>Categories</strong> — 2 endpoints</summary>

| Method | Endpoint | Access |
|---|---|---|
| GET | `/api/v1/categories` | Public |
| GET | `/api/v1/categories/:id` | Public |

</details>

<details>
<summary><strong>Admin</strong> — 21 endpoints</summary>

| Method | Endpoint |
|---|---|
| GET | `/api/v1/admin/users` |
| GET | `/api/v1/admin/users/:id` |
| PATCH | `/api/v1/admin/users/:id` |
| DELETE | `/api/v1/admin/users/:id` |
| GET | `/api/v1/admin/addresses` |
| GET | `/api/v1/admin/addresses/:id` |
| PATCH | `/api/v1/admin/users/:userID/addresses/:id` |
| DELETE | `/api/v1/admin/users/:userID/addresses/:id` |
| GET | `/api/v1/admin/stores` |
| GET | `/api/v1/admin/stores/:id` |
| PATCH | `/api/v1/admin/stores/:id` |
| POST | `/api/v1/admin/stores/:id/reactivate` |
| POST | `/api/v1/admin/stores/:id/deactivate` |
| DELETE | `/api/v1/admin/stores/:id` |
| POST | `/api/v1/admin/categories` |
| POST | `/api/v1/admin/categories/:id/subcategories` |
| PATCH | `/api/v1/admin/categories/:id` |
| PUT | `/api/v1/admin/categories/:id/parent` |
| DELETE | `/api/v1/admin/categories/:id` |
| PATCH | `/api/v1/admin/users/:userID/products/:id` |
| POST | `/api/v1/admin/users/:userID/products/:id/deactivate` |
| POST | `/api/v1/admin/users/:userID/products/:id/reactivate` |
| DELETE | `/api/v1/admin/users/:userID/products/:id` |

</details>

---

## Getting Started

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- Go 1.22+

### Run Locally

```bash
# Clone the repository
git clone https://github.com/yourusername/E-Commerce-API.git
cd E-Commerce-API

# Copy environment variables
cp .env.example .env

# Start all services (PostgreSQL, Redis, Redpanda, MinIO)
docker compose up -d

# Run the API
go run cmd/main.go
```

### Environment Variables

See `.env.example` for all required variables. Key configs:

```env
# Server
APP_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ecommerce
DB_USER=postgres
DB_PASSWORD=

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Redpanda
REDPANDA_BROKERS=localhost:9092

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=
MINIO_SECRET_KEY=

# JWT
JWT_ACCESS_SECRET=
JWT_REFRESH_SECRET=
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Mail
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASSWORD=
```

---

## Project Status

| Feature | Status |
|---|---|
| Auth (register, login, refresh, forgot/reset password) | ✅ Done |
| Email verification via Redpanda worker | ✅ Done |
| Role-based access control (Buyer / Seller / Admin) | ✅ Done |
| User profile & address management | ✅ Done |
| Store management | ✅ Done |
| Product management (soft delete) | ✅ Done |
| Category management (hierarchical) | ✅ Done |
| MinIO — avatar, product image, store logo | 🚧 In Progress |
| Payment gateway abstraction | 🚧 In Progress |
| Real-time chat (WebSocket + Redpanda) | 📋 Planned |
| Audit log & trace ID via Redpanda | 📋 Planned |

---

## License

MIT
