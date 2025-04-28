# User Service (Go)

A Go microservice handling user signup, authentication (JWT), role-based authorization, profile management (including updating a user’s name), and JWT blacklisting, designed for a larger CMS ecosystem.

---

## 1. Objective

- Secure **user signup** (email + password)
- Issue **JWT** tokens on login
- Expose **profile** endpoints (view & update own name)
- Enforce **RBAC** (admin can delete users)
- Support **JWT blacklisting** via Redis
- Expose a **health-check** endpoint

---

## 2. Usage

### Prerequisites

- Docker & Docker Compose
- Go 1.24+
- `make` utility

### Quickstart

1. **Start local databases (Postgres and Redis)**

```bash
docker-compose up -d
```

2. **Apply database migrations**

```bash
make migrate-up
```

3. **Run the user service locally**

```bash
make run
```

4. **Access API**

- Health Check: `GET http://localhost:8080/health`
- Swagger UI: `http://localhost:8080/swagger/index.html`

5. **Run Lint and Tests**

```bash
make lint
make test
```

### Stopping Services

```bash
docker-compose down
```

---

## 3. Tools and Technologies Used

- **Language & Frameworks**:
    - Go 1.24+, Gin (HTTP server)
    - sqlx (Postgres ORM), go-redis (Redis client)
    - Viper (configuration management)
    - testify, mockery (unit testing)
- **Database Migrations**: golang-migrate
- **Documentation**: Swagger via swaggo/gin-swagger
- **Containerization**: Docker, Docker Compose
- **Deployment**: Kubernetes + Istio (for production)
- **Testing**: sqlmock for Postgres, redismock for Redis

---

## 4. Go Get Libraries

```bash
go get github.com/gin-gonic/gin
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get github.com/redis/go-redis/v9
go get github.com/spf13/viper
go get github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go get github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-redis/redismock/v9
go install github.com/vektra/mockery/v3@v3.2.4
go get github.com/golang-jwt/jwt/v5
```

---

## 5. Project Structure, Architecture, and Best Practices

### Project Structure

```
cms-user-service/
├── cmd/server/
│   └── main.go
├── config/
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── internal/
│   ├── auth/
│   ├── cache/
│   ├── config/
│   ├── db/
│   ├── model/
│   ├── repository/
│   ├── service/
│   └── transport/http/
├── migrations/
├── docs/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

### Architecture

- **Hexagonal (Ports and Adapters)**: Transport -> Service -> Repository -> Database/Cache
- **Dependency Injection**: Interfaces are consumer-defined.
- **Viper Config Management**: Environment-specific configurations.
- **Swagger-first Development**: API documentation automatically generated.
- **Migrations Versioning**: Every schema change tracked in Git.
- **Docker-native Development**: Postgres and Redis containerized.
- **Kubernetes-ready Deployment**: Built to deploy into Istio service mesh.

### Best Practices

- Clean layer separation (transport, service, repository)
- Unit tests with mocks for all external dependencies
- Structured logging and health-check endpoints
- Environment-based configuration using Viper
- Secure JWT authentication (secrets managed outside code)

---


