# Golang-todoapp

# 🚀 TodoApp Backend

REST + gRPC backend приложение для управления пользователями и задачами, разработанное на **Go** с использованием принципов **Clean Architecture**.

Проект демонстрирует построение масштабируемого backend-приложения с разделением бизнес-логики, несколькими транспортными слоями (REST и gRPC), PostgreSQL, Docker Compose и структурированным логированием.

---

## 🛠 Tech Stack

<p align="left">
  <img src="https://img.shields.io/badge/Go-1.25.4-00ADD8?style=for-the-badge&logo=go">
  <img src="https://img.shields.io/badge/PostgreSQL-18-4169E1?style=for-the-badge&logo=postgresql">
  <img src="https://img.shields.io/badge/gRPC-Protocol%20Buffers-244C5A?style=for-the-badge&logo=grpc">
  <img src="https://img.shields.io/badge/REST-API-009688?style=for-the-badge">
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker">
  <img src="https://img.shields.io/badge/Zap-Logger-000000?style=for-the-badge">
</p>

---

# ✨ Features

- ✅ User CRUD
- ✅ Task CRUD
- ✅ REST API
- ✅ gRPC API
- ✅ PostgreSQL
- ✅ Docker Compose
- ✅ Graceful Shutdown
- ✅ Context propagation
- ✅ Structured logging (Zap)
- ✅ Clean Architecture
- ✅ Repository Pattern
- ✅ Service Layer
- ✅ Dependency Injection

---

# 🏗 Architecture

```text
                    Client
               (REST / gRPC)
                      │
        ┌─────────────┴─────────────┐
        │      Transport Layer      │
        │   HTTP        │   gRPC    │
        └─────────────┬─────────────┘
                      │
               Service Layer
                      │
             Repository Layer
                      │
                 PostgreSQL
```

The project separates business logic from transport and infrastructure.

- **Transport** — HTTP & gRPC handlers.
- **Service** — business logic.
- **Repository** — database interaction.
- **Core** — shared components (logger, domain models, errors, middleware, server).

---

# 📂 Project Structure

```text
cmd/
└── main.go

internal/
├── core/
│   ├── domain/
│   ├── errors/
│   ├── logger/
│   ├── repository/
│   └── transport/
│
└── features/
    ├── users/
    │   ├── proto/
    │   ├── repository/
    │   ├── service/
    │   └── transport/
    │       ├── http/
    │       └── grpc/
    │
    └── tasks/
    │   ├── repository/
    │   ├── service/
    │   └── transport/
    │       └── http/
    │
    └── stasistics/
        ├── repository/
        ├── service/
        └── transport/
        └── http/

```

---

# 🌐 REST API

Current REST endpoints include:

### Users

```http
POST    /api/v1/users
GET     /api/v1/users
GET     /api/v1/users/{id}
PATCH   /api/v1/users/{id}
DELETE  /api/v1/users/{id}
```

### Tasks

```http
POST    /api/v1/tasks
GET     /api/v1/tasks
GET     /api/v1/tasks/{id}
PATCH   /api/v1/tasks/{id}
DELETE  /api/v1/tasks/{id}
```

---

# ⚡ gRPC API

The project also exposes gRPC endpoints using Protocol Buffers.

Current services:

```protobuf
service UsersService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    rpc GetUsers(GetUsersRequest) returns (GetUsersResponse);
}
```

The gRPC layer reuses the same Service layer as the REST API, allowing multiple transports without duplicating business logic.

---

# 🗄 Database

Database: **PostgreSQL**

The application uses:

- pgx
- SQL migrations
- connection pool
- repository pattern

---

# 🚀 Running the Project

Clone the repository

```bash
git clone <repository_url>
```

Go to the project

```bash
cd Golang-todoapp
```

Create an `.env` file.

Run Docker Compose

```bash
docker compose up --build
```

The application will start together with PostgreSQL.

---

# ⚙ Environment Variables

Example:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=todoapp

DATABASE_HOST=localhost
DATABASE_PORT=5433
```

---

# 📌 Roadmap

Planned improvements:

- JWT Authentication
- Auth Service
- Redis
- Kafka
- Unit Tests
- Integration Tests
- CI/CD
- Prometheus
- Grafana
- OpenTelemetry

---

# 📄 License

This project is distributed under the MIT License.

---

# 👨‍💻 Author

GitHub: **Trykach34rus**

Backend Developer (Go)
