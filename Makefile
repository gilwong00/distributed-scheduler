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

.PHONY: migrateup
migrateup:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" up

.PHONY: migrateuplatest
migrateuplatest:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" up 1

.PHONY: migratedown
migratedown:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" down

.PHONY: migratedownlast
migratedownlast:
	migrate -path internal/taskdb/migrations -database "$(DB_URL)" down 1

.PHONY: proto
proto:
	rm -f proto/gen/*.go
	protoc --proto_path=proto --go_out=proto/gen --go_opt=paths=source_relative \
	--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
	proto/*.proto

.PHONY: startscheduler
startscheduler:
	cd cmd && cd ./scheduler && go run main.go
