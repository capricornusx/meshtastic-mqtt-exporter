BINARY_NAME=meshtastic-exporter
EMBEDDED_BINARY=meshtastic-exporter-embedded

.PHONY: build build-hook clean deps lint test docker run

build: build-standalone build-hook

build-standalone:
	go build -o $(BINARY_NAME) ./cmd/standalone

build-hook:
	go build -o $(EMBEDDED_BINARY) ./cmd/embedded-hook

deps:
	go mod tidy
	go mod download

clean:
	rm -f $(BINARY_NAME) $(EMBEDDED_BINARY)

lint:
	golangci-lint run

test:
	go test -v ./...

test-unit:
	go test -short -v ./...

test-integration:
	go test -run Integration -v ./...

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

coverage-report: coverage
	go tool cover -func=coverage.out

docker:
	docker build -t meshtastic-exporter .

run:
	docker-compose up -d
