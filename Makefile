  .PHONY: help postgres createdb dropdb migrateup migratedown migrate-create migrateup1 migratedown1 sqlc mock  \
  	test test-converage test-integration clean run dev build lint fmt vet deps deps-upgrade \


 DB_URL=postgresql://root:secret@localhost:5433/gomall?sslmode=disable
 DB_CONTAINER=gomall-pg
 DB_NAME=gomall
 DB_USER=root
 DB_PASSWORD=secret

postgres:
	docker run -d \
		--name $(DB_CONTAINER) \
		--network gomall-network \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		postgres:18-alpine
		@echo  "✅ PostgreSQL container started!"

createdb:
	docker exec -it  $(DB_CONTAINER) createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)

dropdb:
	 docker exec -it $(DB_CONTAINER) dropdb --username=$(DB_USER) $(DB_NAME)

migrate-create:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down


migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

migrate-version:
	migrate -path  db/migration -database "$(DB_URL)" version


sqlc:
	sqlc generate

mock:
	 mockgen -package mockdb -destination db/mock/store.go gomall/db/sqlc Querier



test:
	go test -v -cover -short ./...

test-converage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -v -cover ./...


run:
	go run cmd/api/main.go

dev:
	@air

build:
	@go build -o bin/gomall cmd/api/main.go
	@echo "✅ Binary built: bin/gomall"

lint:
	@golangci-lint run

fmt:
	@go fmt ./...
	@echo "✅ Code formatted!"

vet:
	@go vet ./...
	@echo "✅ Code vetted!"

deps:
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated!"

deps-upgrade:
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies upgraded!"
