tidy:
	go mod tidy
	go fmt ./...
	go vet ./...

test:
	go test ./...

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/org_structure_api?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status