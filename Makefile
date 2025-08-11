EMBEDDED_BINARY=embedded-hook
STANDALONE_BINARY=standalone
EXAMPLE_BINARY=mochi-mqtt-integration
TIMEOUT?=30

.PHONY: build build-all build-embedded build-standalone build-example build-examples build-linux build-linux-amd64 build-linux-arm64 clean deps lint test test-unit test-integration coverage docker release-check release-test release-build sonar-up sonar-scan

build: build-all

build-all: build-embedded build-standalone build-example

# Embedded mode with built-in MQTT broker
build-embedded:
	go build -o dist/$(EMBEDDED_BINARY) ./cmd/embedded-hook

# Standalone mode for existing MQTT setups
build-standalone:
	go build -o dist/$(STANDALONE_BINARY) ./cmd/standalone

# Example integration
build-example:
	cd docs/mochi-mqtt-integration && go build -o ../../dist/$(EXAMPLE_BINARY) .

# Alias for CI compatibility
build-examples: build-example


# Dependencies and cleanup
deps:
	go mod tidy
	go mod download

clean:
	rm -f $(EMBEDDED_BINARY) $(STANDALONE_BINARY) $(EXAMPLE_BINARY)
	rm -f coverage.out coverage.html
	rm -rf dist/ reports/

# Code quality
lint:
	golangci-lint run --timeout=5m

# Testing
test: test-unit test-integration coverage-report

test-unit:
	timeout $(TIMEOUT) go test -v ./pkg/...

test-integration:
	timeout $(TIMEOUT) go test -run Integration -v ./tests/...

coverage:
	timeout $(TIMEOUT) go test -race -coverprofile=coverage.out -covermode=atomic ./...
	timeout $(TIMEOUT) go tool cover -html=coverage.out -o coverage.html

coverage-report: coverage
	timeout $(TIMEOUT) go tool cover -func=coverage.out

# Docker
docker:
	docker build -t meshtastic-exporter .

# Cross-compilation
build-linux: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	mkdir -p dist
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(EMBEDDED_BINARY)-linux-amd64 ./cmd/embedded-hook
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(STANDALONE_BINARY)-linux-amd64 ./cmd/standalone

build-linux-arm64:
	mkdir -p dist
	env GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(EMBEDDED_BINARY)-linux-arm64 ./cmd/embedded-hook
	env GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(STANDALONE_BINARY)-linux-arm64 ./cmd/standalone

# Release (requires goreleaser)
release-check:
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install from https://goreleaser.com/install/" && exit 1)
	SSH_PRIVATE_KEY="$$(cat ~/.ssh/github_sign 2>/dev/null || echo '')" goreleaser check

release-test: release-check
	SSH_PRIVATE_KEY="$$(cat ~/.ssh/github_sign 2>/dev/null || echo '')" goreleaser release --snapshot --clean --skip=publish

release-build: release-check
	SSH_PRIVATE_KEY="$$(cat ~/.ssh/github_sign 2>/dev/null || echo '')" goreleaser build --snapshot --clean

# SonarQube
sonar-up:
	@echo "Ожидание запуска SonarQube..."
	@timeout 120 bash -c 'until curl -s http://192.168.1.77:9000/api/system/status | grep -q "UP"; do sleep 5; done' || echo "Таймаут ожидания SonarQube"
	@echo "SonarQube доступен на http://192.168.1.77:9000"

sonar-scan: sonar-up lint-reports coverage
	./scripts/sonar-scan.sh
	@sleep 1
	rm -rf reports/ coverage.out coverage.html

lint-reports:
	@echo "Генерация отчетов для SonarQube..."
	@mkdir -p reports
	@golangci-lint run --out-format checkstyle > reports/golint-report.xml 2>/dev/null || true
	@go vet ./... 2> reports/govet-report.out || true
