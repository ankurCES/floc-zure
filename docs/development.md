# Development Guide

## Setup
```bash
git clone https://github.com/ankurCES/floc-zure.git
cd floc-zure
go mod tidy
make build
```

## Project Structure
```
cmd/azfloci/         Entry point
internal/
  azure/             CLIExecutor interface + implementation
  cli/               Cobra commands (root, version, auth)
  config/            Config Manager interface + Viper impl
  resource/          Resource Manager interface
  workflow/          Workflow Engine interface
pkg/
  models/            Shared types (AzureAccount, Resource, Workflow, Config)
  utils/             Error types (CLIError, AuthError, etc.)
tests/e2e/           End-to-end tests (build tag: e2e)
  helpers/           Test utilities (CLIRunner, assertions, Azure helpers)
  testdata/          Sample YAML fixtures
```

## Running Tests
```bash
make test          # Unit tests (no Azure needed)
make e2e-test      # E2E tests (needs az login)
make lint          # golangci-lint
make vet           # go vet
```

## Adding a New Command
1. Create `internal/cli/mycommand.go`
2. Define `var myCmd = &cobra.Command{...}`
3. Register in `init()`: `rootCmd.AddCommand(myCmd)`
4. If it calls Azure, use `azure.NewCLIExecutorImpl()`

## Coding Conventions
- **Interfaces** in `internal/*/` (e.g., `manager.go`), implementations in `*_impl.go`
- **Models** in `pkg/models/types.go` — shared across packages
- **Errors** in `pkg/utils/errors.go` — typed errors with context
- **Tests**: table-driven, mock `CLIExecutor` for unit tests
- **Formatting**: `gofmt`, conventional commits (`feat:`, `fix:`, `docs:`)

## Mock Testing Pattern
```go
type mockExecutor struct{}
func (m *mockExecutor) Run(ctx context.Context, args ...string) (*azure.CLIResult, error) {
    return &azure.CLIResult{Stdout: `{"id":"test"}`, ExitCode: 0}, nil
}
// Use in tests instead of shelling out to az
```

## Release
```bash
git tag v0.1.0
goreleaser release --clean
```
