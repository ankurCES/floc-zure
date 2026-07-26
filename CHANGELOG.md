# Changelog

All notable changes to azfloci are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.2.0] — 2025-07-26

### Added
- **Storage Account Simulation**: Full CRUD for storage accounts, blob containers, and blobs (`az storage account/container/blob create/show/list/delete`). Cascade deletes propagate from account → containers → blobs.
- **Key Vault Simulation**: Full CRUD for vaults, secrets, and keys (`az keyvault create/show/list/delete`, `az keyvault secret set/show/list/delete`, `az keyvault key create/show/list/delete`). Secret versioning support.
- **CI/CD Pipeline**: GitHub Actions workflow with matrix build (Go 1.22/1.23 × ubuntu-latest/macos-latest), golangci-lint, unit tests, build verification, and simulator E2E tests.
- **Makefile targets**: `lint-ci` (GitHub Actions output format), `test-sim-unit` (simulator unit tests only), `test-all` (combined unit + simulator tests).
- **Updated simulator help**: `--help` / `--version` now lists all 38 simulated commands including storage and keyvault.
- **Updated simulator design doc**: Storage and keyvault commands added to command table and state schema.

### Fixed
- **Workflow engine race condition**: Fixed data race in concurrent step execution (shared `callOrder` slice now guarded by mutex).
- **Empty state file handling**: Simulator store now treats empty state files as fresh (seeds default data instead of failing to parse).

### Changed
- **README.md**: Complete rewrite with SVG logo, 10 badge pills, full simulator command reference, updated architecture diagram, workflow examples, and documentation links.
- **ROADMAP.md**: Phases 6–8 marked complete.

## [0.1.0] — 2025-07-25

### Added
- **CLI framework**: Cobra-based command tree with `version`, `auth status` commands.
- **Azure CLI executor**: `CLIExecutor` interface wrapping `az` via `os/exec`. Supports `Run()`, `RunJSON()`, `GetAccount()`, `IsAuthenticated()`, `SetSubscription()`.
- **Configuration management**: Viper-based `config.Manager` with YAML file, env vars (`AZFLOCI_*`), and defaults.
- **Resource management**: `azfloci group create/list/show/delete`, `azfloci resource list/show/delete/tag`.
- **Workflow engine**: YAML pipeline execution with DAG dependencies, topological sort, concurrent independent step execution, error policies (`continue`/`abort`/`retry`).
- **Config CLI**: `azfloci config set/get/list/init`.
- **Azure Cloud Simulator**: Drop-in fake `az` binary, JSON-backed state store, account/group/resource handlers, auto-seeded test subscription.
- **Data models**: `AzureAccount`, `ResourceGroup`, `Resource`, `Workflow`, `WorkflowStep`, `StepResult`, `WorkflowResult`, `Config`.
- **Error types**: `CLIError`, `AuthError`, `ValidationError`, `WorkflowError` (with `Unwrap()`).
- **Unit tests**: 76+ tests across 5 packages. Simulator tests: 38+ across state/router/handlers.
- **E2E test suite**: 22 tests covering auth, config, resources, workflows, error handling.
- **Build system**: Makefile with `build`, `test`, `e2e-test`, `lint`, `fmt`, `vet`, `sim-build`, `sim-test` targets.
- **GoReleaser config**: Cross-compilation for linux/darwin/windows × amd64/arm64.
- **Documentation**: Architecture doc, command reference, getting started guide, Azure setup guide, migration guide, development guide, simulator design doc.
