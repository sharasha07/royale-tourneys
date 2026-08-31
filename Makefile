## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	go fmt ./...
	go vet ./...
	staticcheck ./...
	CGO_ENABLED=1 go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	go mod tidy
	go mod verify
	go mod vendor

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	go build -ldflags='-s' -o=./bin/api ./cmd/api
