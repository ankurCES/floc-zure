# Contributing to azfloci

## Quick Start
1. Fork & clone
2. `go mod tidy && make build`
3. Create a branch: `git checkout -b feat/my-feature`
4. Make changes, add tests
5. `make test && make lint`
6. Open a PR

## Code Style
- `gofmt` / `golangci-lint`
- Interfaces in `internal/*/`, models in `pkg/models/`
- Table-driven tests, mock the `CLIExecutor` interface

## Commit Messages
Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `chore:`
