include .env

MIGRATE_PATH = "pkg/database/migrations"
MIGRATE_DB = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)"

run:
	@go run cmd/api/main.go

generate:
	@go generate ./...

test-logic:
	@go test ./internal/features/*/service -cover

#----------------- Start Docker -----------
# -----------------------------------------
# build db & app
docker-up:
	@docker compose up -d --build

# Dev Mode: build db only
docker-db:
	@docker compose up -d db

# clear volumns data
docker-clear:
	@docker compose down -v

docker-config:
	@docker compose config


#----------------- Start Migrations -----------
# ---------------------------------------------
# migrate create usage: make migrate-create name=init_example
migrate-create:
	@migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

migrate-up:
	@migrate -database $(MIGRATE_DB) -path $(MIGRATE_PATH) up

# down 1 step for fix error
migrate-down:
	@migrate -database $(MIGRATE_DB) -path $(MIGRATE_PATH) down 1

# migrate force usage: make migrate-force version=1
migrate-force:
	@migrate -database $(MIGRATE_DB) -path $(MIGRATE_PATH) force $(version)

# NOTE: reset all migrations
migrate-reset:
	@migrate -database $(MIGRATE_DB) -path $(MIGRATE_PATH) down
