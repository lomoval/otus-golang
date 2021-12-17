BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

# Postgres - for migrations
PG_HOST := $(if ${POSTGRES_HOST},${POSTGRES_HOST},127.0.0.1)
PG_PORT := $(if ${POSTGRES_PORT},${POSTGRES_PORT},5432)
PG_USER := $(if ${POSTGRES_USERNAME},${POSTGRES_USERNAME},postgres)
PG_PASS := $(if ${POSTGRES_PASSWORD},${POSTGRES_PASSWORD},pas)

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

test-all:
	go test -v --tags sql -race ./internal/...

migrate:
	 goose postgres "dbname=postgres sslmode=disable host=${PG_HOST} port=${PG_PORT} password=${PG_PASS} user=${PG_USER}" up

migrate-testing:
	 goose postgres "dbname=testing sslmode=disable host=${PG_HOST} port=${PG_PORT} password=${PG_PASS} user=${PG_USER}" up

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint
