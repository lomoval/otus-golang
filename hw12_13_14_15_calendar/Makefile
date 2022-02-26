BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
SCHEDULER_BIN := "./bin/calendar_scheduler"
SCHEDULER_DOCKER_IMG="scheduler:develop"
SENDER_BIN := "./bin/calendar_sender"
SENDER_DOCKER_IMG="sender:develop"
MIGRATIONS_DOCKER_IMG="migrations:develop"
INTEGRATION_TEST_DOCKER_IMG="integration-test:develop"


# Postgres - for migrations
PG_HOST := $(if ${POSTGRES_HOST},${POSTGRES_HOST},127.0.0.1)
PG_PORT := $(if ${POSTGRES_PORT},${POSTGRES_PORT},5432)
PG_USER := $(if ${POSTGRES_USERNAME},${POSTGRES_USERNAME},postgres)
PG_PASS := $(if ${POSTGRES_PASSWORD},${POSTGRES_PASSWORD},pas)

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/scheduler
	go build -v -o $(SENDER_BIN) -ldflags "$(LDFLAGS)" ./cmd/sender

run: build
	$(BIN) -config ./configs/config.yaml

build-img-all: build-img build-img-sender build-img-scheduler build-img-migrations build-img-integration-tests

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .


build-img-sender:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(SENDER_DOCKER_IMG) \
		-f build/Dockerfile-sender .

build-img-scheduler:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(SCHEDULER_DOCKER_IMG) \
		-f build/Dockerfile-scheduler .

build-img-migrations:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(MIGRATIONS_DOCKER_IMG) \
		-f build/Dockerfile-migrations .

build-img-integration-tests:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(INTEGRATION_TEST_DOCKER_IMG) \
		-f build/Dockerfile-integration-test .

integration-tests:
	docker-compose -f build/docker-compose.yaml -f build/docker-compose-integration-test.yaml run migrations
	docker-compose -f build/docker-compose.yaml -f build/docker-compose-integration-test.yaml run integration_tests

run-img: build-img
	docker run $(DOCKER_IMG)

up:
	docker-compose -f build/docker-compose.yaml up
# 	docker-compose run postgres goose postgres "dbname=postgres sslmode=disable host=${PG_HOST} port=${PG_PORT} password=${PG_PASS} user=${PG_USER}" up

down:
	docker-compose -f build/docker-compose.yaml down

version: build
	$(BIN) version

test:
	go test -race ./...

test-all:
	go test --tags sql -race ./...

test-all-clean-cache:
	go clean -testcache
	go test --tags sql -race ./...

migrate:
	 goose postgres "dbname=postgres sslmode=disable host=${PG_HOST} port=${PG_PORT} password=${PG_PASS} user=${PG_USER}" up

migrate-testing:
	 goose postgres "dbname=testing sslmode=disable host=${PG_HOST} port=${PG_PORT} password=${PG_PASS} user=${PG_USER}" up

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

lint-fix: install-lint-deps
	golangci-lint run ./... --fix

install-gen-deps:
	(which protoc-gen-go > /dev/null) || go install google.golang.org/protobuf/cmd/protoc-gen-go
	(which protoc-gen-go-grpc > /dev/null) || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	(which protoc-gen-grpc-gateway > /dev/null) || go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	(which protoc-gen-openapiv2 > /dev/null) || go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

generate: install-gen-deps
	go generate ./...


#go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: build run build-img run-img version test lint
