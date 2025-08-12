#!/bin/env bash

HOOK_FILE=".git/hooks/pre-commit"

cat > "$HOOK_FILE" << 'EOF'
#!/bin/bash

set -e

echo "Running pre-commit checks..."

# Run linter
echo "Running golangci-lint..."
make lint

# Run tests
echo "Running tests..."
make test

# Build binaries
echo "Building binaries..."
make build

# Check GoReleaser configuration if modified
if git diff --cached --name-only | grep -q "\.goreleaser\.yaml"; then
    echo "Checking GoReleaser configuration..."
    if command -v goreleaser &> /dev/null; then
        goreleaser check
    else
        echo "Warning: goreleaser not found, skipping configuration check"
    fi
fi

echo "All pre-commit checks passed!"
EOF

chmod +x "$HOOK_FILE"
echo "Pre-commit hook installed successfully!"
