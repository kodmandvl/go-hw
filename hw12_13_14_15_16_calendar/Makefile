BIN := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
DOCKER_IMG_CALENDAR ?= calendar-api:develop
DOCKER_IMG_SCHEDULER ?= calendar-scheduler:develop
DOCKER_IMG_SENDER ?= calendar-sender:develop
DOCKER_LOGIN ?= kodmandvl
DOCKER_REGISTRY ?= docker.io
DOCKER_PROJECT ?= $(DOCKER_REGISTRY)/$(DOCKER_LOGIN)
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Software to run containers
CONTAINER_RUNNER ?= docker # docker|podman
COMPOSE ?= $(CONTAINER_RUNNER) compose
COMPOSE_FILE ?= ./deployments/docker-compose.yaml
COMPOSE_FILE_IT ?= ./deployments/docker-compose.integration.yaml

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-calendar
POSTGRES_REGISTRY ?= docker.io/library/postgres
POSTGRES_TAG ?= latest # latest or 16.13-alpine3.22 or your other desirable tag

#Rabbit
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq
RABBIT_REGISTRY ?= docker.io/library/rabbitmq
RABBIT_TAG ?= management

# for test:
# podman run -d --name postgres-calendar --hostname postgres-calendar -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=backend -p 127.0.0.1:5435:5432 docker.io/library/postgres:16.13-alpine3.22
# psql postgres://postgres:password@localhost:5435/backend
run-postgres:
	$(CONTAINER_RUNNER) run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql/data \
	$(POSTGRES_REGISTRY):$(POSTGRES_TAG)

# for test:
# podman run -d --name rabbitmq --hostname rabbitmq -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=rabbit -e RABBITMQ_DEFAULT_PASS=password docker.io/library/rabbitmq:management
run-rabbit:
	$(CONTAINER_RUNNER) run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	$(RABBIT_REGISTRY):$(RABBIT_TAG)

build-calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

build-sender:
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

build: build-calendar build-scheduler build-sender

run: build-calendar
	$(BIN) --config ./configs/calendar_config.yaml

# run-calendar - то же, что и просто run (просто run как бы для обратной совместимости):
run-calendar: build-calendar
	$(BIN) --config ./configs/calendar_config.yaml

run-scheduler: build-scheduler
	$(BIN_SCHEDULER) --config ./configs/scheduler_config.yaml

run-sender: build-sender
	$(BIN_SENDER) --config ./configs/sender_config.yaml

build-img-calendar:
	$(CONTAINER_RUNNER) build \
		-t $(DOCKER_PROJECT)/$(DOCKER_IMG_CALENDAR) \
		-f build/Dockerfile.calendar .

build-img-scheduler:
	$(CONTAINER_RUNNER) build \
		-t $(DOCKER_PROJECT)/$(DOCKER_IMG_SCHEDULER) \
		-f build/Dockerfile.scheduler .

build-img-sender:
	$(CONTAINER_RUNNER) build \
		-t $(DOCKER_PROJECT)/$(DOCKER_IMG_SENDER) \
		-f build/Dockerfile.sender .

build-img: build-img-calendar build-img-scheduler build-img-sender

# run-img - только контейнер с календарём
run-img: build-img-calendar
	$(CONTAINER_RUNNER) run $(DOCKER_PROJECT)/$(DOCKER_IMG_CALENDAR)

version: build-calendar
	$(BIN) version

# version-calendar - то же, что и просто version (просто version как бы для обратной совместимости):
version-calendar: build-calendar
	$(BIN) version

version-scheduler: build-scheduler
	$(BIN_SCHEDULER) version

version-sender: build-sender
	$(BIN_SENDER) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.63.4

lint: install-lint-deps
	golangci-lint run ./...

generate:
	go generate ./...

up:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build

down:
	$(COMPOSE) -f $(COMPOSE_FILE) down -v --remove-orphans

integration-tests:
	@set -e; \
	trap '$(COMPOSE) -f $(COMPOSE_FILE_IT) down -v --remove-orphans' EXIT; \
	$(COMPOSE) -f $(COMPOSE_FILE_IT) up -d --build postgres rabbitmq calendar scheduler sender; \
	$(COMPOSE) -f $(COMPOSE_FILE_IT) build integration-tests; \
	$(COMPOSE) -f $(COMPOSE_FILE_IT) run --rm integration-tests

.PHONY: build run-postgres run-rabbit build-calendar build-scheduler build-sender build run run-calendar run-scheduler run-sender build-img build-img-calendar build-img-scheduler build-img-sender run-img version version-calendar version-scheduler version-sender test install-lint-deps lint generate up down integration-tests
