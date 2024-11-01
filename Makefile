DB_URL=postgres://postgres:postgres@localhost:5432/task?sslmode=disable

# HELP
# This will output the help for each task
.PHONY: help
help: ## List all commands and sub-commands in the Makefile.
	@echo '[Lummis-Server] Available Make Targets:'
	@awk '/^[a-zA-Z_0-9.-]+:/ { print $$1 }' $(MAKEFILE_LIST) | sed 's/:$$//' | sort | uniq

.PHONY: setup
setup: ## Install required dependencies
	go mod tidy
	brew install sqlc
	brew install golang-migrate

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: migration
migration:
	migrate create -ext sql -dir internal/taskdb/migrations "$(filter-out $@,$(MAKECMDGOALS))"

.PHONY: dockerup
dockerup:
		docker compose --verbose -p task -f ./docker-compose.yml up -d

migrateup:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" up

migrateuplatest:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" up 1

migratedown:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" down

migratedownlast:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" down 1


.PHONY: startscheduler
startscheduler:
	cd cmd && cd ./scheduler && go run main.go
