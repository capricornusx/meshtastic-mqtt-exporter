# Contributing

## Development Setup

1. Clone the repository
2. Install Go 1.24+
3. Install golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
4. Run `make deps` to install dependencies
5. Run `make build` to build all versions

## Testing

```bash
make lint  # Run linter
make test  # Run tests
```

- All tests use `testify` for assertions
- Tests should be parallel `t.Parallel()`
- A table-driven approach is used for multiple scenarios
- Mocks are located in `pkg/mocks/`


## Code Style

- Use `gofmt` for formatting
- Follow Go naming conventions
- Keep functions small and focused
- Add comments for exported functions

## Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

