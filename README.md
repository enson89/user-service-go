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

- Health Check: `GET http://localhost:8080/v1/health`
- Swagger UI: `http://localhost:8080/v1/swagger/index.html`

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
├── .dockerignore
├── .gitgnore
├── .golangci.yml
├── .mockery.yml
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

# CI/CD Pipeline Overview

Our GitHub Actions workflow automates every stage of code delivery: quality checks, security scans, testing, container builds, and deployment. Below is an explanation of each phase and why it matters.

---

## Triggers

- **Push to `main`**  
  Every commit merged into the `main` branch kicks off the full pipeline up through image build & push. This guarantees that only validated code ever results in a container image.

- **Manual dispatch**  
  A separate “Deploy” job can be triggered by hand (via the **Run workflow** button). This decouples image creation from production rollout, giving operators control over when a release actually goes live.

---

## Environment & Secrets

Before the jobs run, we rely on a handful of repository-level secrets and environment variables:

- **AWS credentials** (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`): allow the runner to log in to ECR and talk to EKS.
- **`AWS_REGION`**: ensures all AWS CLI commands target the correct region.
- **`AWS_ECR_REPOSITORY`**: the name of the ECR repo where we push Docker images.
- **`CLUSTER_NAME`**: the EKS cluster identifier used when updating kubeconfig for `kubectl`.

Keeping these out of code and in GitHub Secrets enforces good security hygiene.

---

## 1. Lint Job

**Purpose:** Enforce coding standards and catch errors at the earliest stage.

- **What happens:**
  - The job checks out your code and sets up Go.
  - It restores a cache of downloaded Go modules to speed up the run.
  - It runs `golangci-lint run` with a timeout to ensure all enabled linters (e.g., `errcheck`, `staticcheck`, `gofmt`) pass before moving on.

**Why it matters:**  
Early lint failures prevent poorly styled or obviously broken code from progressing further.

---

## 2. Vulnerability Scan

**Purpose:** Detect any known security vulnerabilities in your dependencies.

- **What happens:**
  - Again, code is checked out and Go is configured.
  - The same Go modules cache is restored.
  - `govulncheck` examines all imported packages and flags any versions with known CVEs.

**Why it matters:**  
Catching vulnerable libraries before deployment reduces your attack surface and ensures compliance with security policies.

---

## 3. Test & Coverage

**Purpose:** Verify that all logic behaves as expected, and measure test coverage.

- **What happens:**
  - After checkout and Go setup, the Go build cache is restored.
  - `go vet` runs static analysis for suspicious patterns.
  - Tests are executed across all packages **except** the generated mocks directory, producing a `coverage.out` report.
  - The coverage artifact is uploaded so you can inspect test completeness.

**Why it matters:**  
Automated testing prevents regressions and gives confidence in code correctness. Coverage metrics ensure that critical paths aren’t untested.

---

## 4. Build & Push Docker Image

**Purpose:** Package the service into a container and publish it to Amazon ECR.

- **What happens:**
  1. AWS credentials are configured so the runner can authenticate.
  2. The `aws-actions/amazon-ecr-login` action obtains an ECR login and exposes the `registry` URL.
  3. Docker Buildx is set up, enabling cross-platform and cached builds.
  4. Layers are cached using GitHub’s cache backend (`type=gha`) to avoid rebuilding unchanged steps.
  5. The image is built for `linux/amd64`, tagged with the Git SHA, and pushed to the specified ECR repository.
  6. Provenance (SBOM/attestation) is disabled for faster runtimes—only the raw image is published.

**Why it matters:**  
Automated, cached builds maximize speed and consistency. Tagging by SHA ensures immutability, and pushing to ECR readies the image for deployment.

---

## 5. Deploy to EKS (Manual)

**Purpose:** Roll out the newly built image to the Kubernetes cluster on demand.

- **What happens:**
  - The job runs only when you manually trigger it.
  - AWS CLI updates the local kubeconfig so `kubectl` can talk to your EKS cluster.
  - A single `kubectl set image` command updates the `user-service-go` Deployment to use the new image tag.
  - `kubectl rollout status` waits for the deployment to complete successfully before marking the job done.

**Why it matters:**  
Manual approval for deployment adds a safety gate. You can validate the new image (e.g., in a staging environment) before promoting it to production.

---

## Caching Strategy

1. **Go Modules & Build Cache**
  - Restored at the start of each Go-related job to avoid re-downloading dependencies and recompiling everything from scratch.

2. **Docker Layers**
  - Leveraging Buildx’s cache with GitHub Actions (`cache-from`/`cache-to`) ensures only changed layers are rebuilt, speeding up CI.

Caching dramatically reduces pipeline run times and conserves CI resources.

---

## Extensibility

- **Semantic versioning:** You can enhance the pipeline to tag images with `vX.Y.Z` by reading Git tags or auto-bumping on merges.
- **Multi-arch builds:** Add other `platforms:` entries to support ARM or other architectures.
- **Approval gates:** Insert manual review steps or environment protections in the GitHub Actions UI before deployment.

This modular design makes it easy to evolve the pipeline as your project grows.  


