## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	dotenvx run -- go run ./cmd/api

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	dotenvx run -- sh -c 'psql $${POSTGRES_URL}'

## db/migrate/create name=$1: create a new database migration
.PHONY: db/migrate/create
db/migrate/create:
	migrate create -ext=.sql -dir=./migrations -seq ${name}

## db/migrate/up: apply all up database migrations
.PHONY: db/migrate/up
db/migrate/up:
	dotenvx run -- sh -c 'migrate -path=./migrations -database=$${POSTGRES_URL} up'

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit:
	go mod tidy
	go mod verify
	go fmt ./...
	go vet ./...
	staticcheck ./...
	CGO_ENABLED=1 go test -race -vet=off ./...

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	go build -ldflags='-s' -o=./bin/api ./cmd/api
