#!/bin/bash

HOOK_FILE=".git/hooks/pre-commit"

cat > "$HOOK_FILE" << 'EOF'
#!/bin/bash

set -e

echo "Running pre-commit checks..."

# Run linter
echo "Running golangci-lint..."
golangci-lint run

# Run tests
echo "Running tests..."
go test -race ./...

# Build binaries
echo "Building binaries..."
go build -o /tmp/meshtastic-exporter ./cmd/standalone
go build -o /tmp/meshtastic-exporter-embedded ./cmd/embedded-hook

echo "All pre-commit checks passed!"
EOF

chmod +x "$HOOK_FILE"
echo "Pre-commit hook installed successfully!"