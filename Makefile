BIN := "./bin/minipic"
DOCKER_IMG="minipic"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S%:z) -X main.gitHash=$(GIT_HASH)

build:
	go build -o bin/minipic -ldflags "$(LDFLAGS)" cmd/*

run: build
	./bin/minipic --config=configs/config.toml

tests:
	go test -v -count=20 -race -timeout=10m ./internal/app/...
	go test -v -count=5 -race -timeout=15m ./test/...

build-image:
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG) -f build/Dockerfile .

run-image: build-image
	docker run -p 9011:9011 --rm $(DOCKER_IMG)

install-lint:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.37.0

lint: install-lint
	golangci-lint run

.PHONY: build run tests build-image run-image install-lint lint