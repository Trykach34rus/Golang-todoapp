# Golang-todoapp

# 🚀 TodoApp Backend

REST + gRPC backend приложение для управления пользователями и задачами, разработанное на **Go** с использованием принципов **Clean Architecture**.

Проект демонстрирует построение масштабируемого backend-приложения с разделением бизнес-логики, несколькими транспортными слоями (REST и gRPC), PostgreSQL, Docker Compose и структурированным логированием.

---

# 🛠 Tech Stack

* **Go**
* **PostgreSQL**
* **pgx / pgxpool**
* **REST API**
* **gRPC**
* **Protocol Buffers**
* **Docker / Docker Compose**
* **Zap**
* **Swagger / OpenAPI**
* **Clean Architecture**

---

# ✨ Features

* ✅ User CRUD
* ✅ Task CRUD
* ✅ REST API
* ✅ gRPC API
* ✅ PostgreSQL
* ✅ Docker Compose
* ✅ Graceful Shutdown
* ✅ Context propagation
* ✅ Structured logging (Zap)
* ✅ Clean Architecture
* ✅ Repository Pattern
* ✅ Service Layer
* ✅ Dependency Injection
* ✅ Swagger / OpenAPI documentation

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

* **Transport** — HTTP & gRPC handlers.
* **Service** — business logic.
* **Repository** — database interaction.
* **Core** — shared components (logger, domain models, errors, middleware, server).

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
    ├── tasks/
    │   ├── repository/
    │   ├── service/
    │   └── transport/
    │       └── http/
    │
    └── statistics/
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

# 📚 Swagger API Documentation

The REST API is documented using **Swagger / OpenAPI**.

You can open the interactive Swagger UI here:

[Swagger UI — TodoApp API](http://84.38.180.116:5050/swagger/index.html?utm_source=chatgpt.com#/tasks/get_tasks)

Swagger UI allows you to:

* view all available REST endpoints;
* inspect request and response models;
* see HTTP methods and parameters;
* send test requests directly to the running API;
* inspect API responses.

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

* pgx
* SQL migrations
* connection pool
* repository pattern

---

# 🚀 Running the Project

Clone the repository:

```bash
git clone https://github.com/Trykach34rus/Golang-todoapp.git
```

Go to the project:

```bash
cd Golang-todoapp
```

Create an `.env` file.

Run Docker Compose:

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

# 📄 License

This project is distributed under the MIT License.

---

# 👨‍💻 Author

GitHub: **Trykach34rus**

Backend Developer (Go)
