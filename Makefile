EMBEDDED_BINARY=embedded-hook
STANDALONE_BINARY=standalone
EXAMPLE_BINARY=mochi-mqtt-integration
TIMEOUT?=120

.PHONY: build build-all build-embedded build-standalone build-example build-examples build-linux build-linux-amd64 build-linux-arm64 clean deps lint test test-unit test-integration test-e2e coverage docker release-check release-test release-build sonar-up sonar-scan

build: build-all

build-all: build-embedded build-standalone build-example

# Embedded mode with built-in MQTT broker
build-embedded:
	@CGO_ENABLED=0 go build -ldflags="-X meshtastic-exporter/pkg/version.Version=dev -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(EMBEDDED_BINARY) ./cmd/embedded-hook

# Standalone mode for existing MQTT setups
build-standalone:
	@CGO_ENABLED=0 go build -ldflags="-X meshtastic-exporter/pkg/version.Version=dev -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(STANDALONE_BINARY) ./cmd/standalone

# Example integration
build-example:
	@cd docs/mochi-mqtt-integration && go build -o ../../dist/$(EXAMPLE_BINARY) .

build-examples: build-example


# Dependencies and cleanup
deps:
	go mod tidy
	go mod download

clean:
	rm -f $(EMBEDDED_BINARY) $(STANDALONE_BINARY) $(EXAMPLE_BINARY)
	rm -f coverage.out
	rm -rf dist/ reports/

lint:
	golangci-lint run --timeout=5m

test: test-unit test-integration test-e2e coverage-report
	@rm -f meshtastic_state.json

test-unit:
	@timeout $(TIMEOUT) go test -parallel 4 ./pkg/...

test-integration:
	@timeout $(TIMEOUT) go test -run Integration -parallel 4 ./tests/...

test-e2e:
	@timeout $(TIMEOUT) go test -run TestE2E ./tests/...

coverage:
	@timeout $(TIMEOUT) go test -race -parallel 4 -coverprofile=coverage.out -covermode=atomic ./... 2>/dev/null
	@timeout $(TIMEOUT) go tool cover -func=coverage.out 2>/dev/null

coverage-report: coverage

docker:
	docker build -t meshtastic-exporter .

# Cross-compilation
build-linux: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	@mkdir -p dist
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X meshtastic-exporter/pkg/version.Version=$$(git describe --tags --always 2>/dev/null || echo dev) -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(EMBEDDED_BINARY)-linux-amd64 ./cmd/embedded-hook
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X meshtastic-exporter/pkg/version.Version=$$(git describe --tags --always 2>/dev/null || echo dev) -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(STANDALONE_BINARY)-linux-amd64 ./cmd/standalone

build-linux-arm64:
	@mkdir -p dist
	@env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X meshtastic-exporter/pkg/version.Version=$$(git describe --tags --always 2>/dev/null || echo dev) -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(EMBEDDED_BINARY)-linux-arm64 ./cmd/embedded-hook
	@env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X meshtastic-exporter/pkg/version.Version=$$(git describe --tags --always 2>/dev/null || echo dev) -X meshtastic-exporter/pkg/version.GitCommit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X meshtastic-exporter/pkg/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/$(STANDALONE_BINARY)-linux-arm64 ./cmd/standalone

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
	@rm -rf reports/ coverage.out

lint-reports:
	@echo "Генерация отчетов для SonarQube..."
	@mkdir -p reports
	@golangci-lint run --out-format checkstyle > reports/golint-report.xml 2>/dev/null || true
	@go vet ./... 2> reports/govet-report.out || true
