# CLAUDE.md - Movie Terminal API

## Build and Test Commands
- **Build**: `go build -v ./...`
- **Test**: `go test -v ./...`
- **Lint**: `golangci-lint run`
- **Install Deps**: `go mod tidy`

## Code Style & Guidelines
- **Language**: Go 1.24+
- **Linter**: Use `golangci-lint`. Enforce `nlreturn` (blank line before return).
- **Formatting**: Always run `gofmt` before committing.
- **Naming**: Use camelCase for internal variables and PascalCase for exported types.
- **Errors**: Wrap errors with context using `fmt.Errorf("context: %w", err)`.
- **Architecture**: Keep the API handlers separate from the business logic in `internal/`.