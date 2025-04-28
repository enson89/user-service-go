.PHONY: lint test run migrate-up migrate-down migrate-status migrate-goto migrate-drop migrate-create

# ── Configuration ─────────────────────────────────────────────────────────────

# Choose which section in config.yml to load
ENV ?= dev

# Path to the per‐env YAML
CONFIG_FILE := internal/config/config.$(ENV).yml

# Migrate CLI (install via: go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
MIGRATE := migrate

# Where your SQL migrations live
MIGRATIONS_DIR    := ./migrations
MIGRATIONS_SOURCE := file://$(MIGRATIONS_DIR)

# Pull nested DB settings using yq
DB_USER     := $(shell yq e '.db.user'     $(CONFIG_FILE))
DB_PASSWORD := $(shell yq e '.db.password' $(CONFIG_FILE))
DB_HOST     := $(shell yq e '.db.host'     $(CONFIG_FILE))
DB_PORT     := $(shell yq e '.db.port'     $(CONFIG_FILE))
DB_NAME     := $(shell yq e '.db.name'     $(CONFIG_FILE))
DB_SSLMODE  := $(shell yq e '.db.sslmode'  $(CONFIG_FILE))

# Compose the full Postgres URL
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Enable verbose logging by passing VERBOSE=1
ifeq ($(VERBOSE),1)
  MIGRATE_VERBOSE := -verbose
endif

# ── Targets ────────────────────────────────────────────────────────────────────

## lint: static analysis
lint:
	@echo "🔍 Running go vet & golangci-lint…"
	go vet ./...
	golangci-lint run

## test: generate mocks & run coverage
PKGS := $(shell go list ./... | grep -v '/mocks')
test:
	@echo "✅ Running tests…"
	mockery
	go test $(PKGS) -cover

## run: start your app in dev
run:
	@echo "🌱 Starting app in $(ENV) mode…"
	USER_SVC_ENV=$(ENV) go run cmd/api/main.go

swagger:
	@echo "📝 Generating Swagger docs…"
	swag init --generalInfo cmd/api/main.go --output docs

## migrate-up: apply all pending up migrations
migrate-up:
	@echo "▶️  $(ENV): Applying migrations (up)…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  -source $(MIGRATIONS_SOURCE) \
	  -database "$(DB_URL)" up

## migrate-down: roll back the last migration
migrate-down:
	@echo "↩️  $(ENV): Rolling back one migration…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  -source $(MIGRATIONS_SOURCE) \
	  -database "$(DB_URL)" down 1

## migrate-status: show current schema version
migrate-status:
	@echo "ℹ️  $(ENV): Migration status/version…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  -source $(MIGRATIONS_SOURCE) \
	  -database "$(DB_URL)" version

## migrate-goto: migrate to a specific version
# Usage: make migrate-goto VERSION=42
migrate-goto:
ifndef VERSION
	$(error VERSION is required! Usage: make migrate-goto VERSION=<n>)
endif
	@echo "⏩  $(ENV): Migrating to version $(VERSION)…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  -source $(MIGRATIONS_SOURCE) \
	  -database "$(DB_URL)" goto $(VERSION)

## migrate-drop: drop all objects in the DB
migrate-drop:
	@echo "💣  $(ENV): Dropping all database objects…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  -source $(MIGRATIONS_SOURCE) \
	  -database "$(DB_URL)" drop -f

## migrate-create: scaffold a new timestamped up/down pair
# Usage: make migrate-create NAME=add_users_table
migrate-create:
ifndef NAME
	$(error NAME is required! Usage: make migrate-create NAME=<migration_name>)
endif
	@echo "✏️  $(ENV): Creating new migration $(NAME)…"
	$(MIGRATE) $(MIGRATE_VERBOSE) \
	  create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)