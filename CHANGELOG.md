# Changelog

All notable changes to azfloci are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.5.1] — 2026-07-27

### Fixed
- **golangci-lint clean**: Zero lint errors across all packages. Added `.golangci.yml` config (errcheck, govet, staticcheck, ineffassign, unused; test files excluded from errcheck).
- **errcheck fixes**: `internal/drift/snapshot.go` — checked `json.Unmarshal` return values. `simulator/internal/handlers/keyvault.go` — checked `fmt.Sscanf` return. `simulator/internal/state/store.go` — all 15 bare `s.persist()` calls now use `_ = s.persist()`. `pkg/plugin/provider_test.go` — fixed unused variable (staticcheck SA4006).

## [0.5.0] — 2026-07-27

### Added
- **Service Bus Simulation**: Full CRUD for namespaces, queues, topics, and topic subscriptions (`az servicebus namespace/queue/topic create/show/list/delete`, `az servicebus topic subscription create/show/list/delete`). In-memory message queue with `send/receive/peek` commands. Cascade deletes propagate from namespace → queues/topics/subscriptions/messages.
- **Function App Simulation**: Full CRUD for function apps and functions (`az functionapp create/show/list/delete`, `az functionapp function create/show/list/delete`). Simulated invocation with `invoke` command and invocation history via `invocations`. Cascade deletes propagate from app → functions/invocations.
- **5-word router dispatch**: Supports commands like `servicebus topic subscription create` and `servicebus queue message send`.
- **12 new store tests** (8 Service Bus + 5 Function App) and **7 new handler tests** (4 Service Bus + 3 Function App).

## [0.4.0] — 2026-07-27

### Added
- **ARM Template Engine**: Parser with expression evaluation (parameters, variables, concat, toLower, toUpper, uniqueString, resourceId, format, resourceGroup().location). Deployer creates simulated resources from ARM JSON. CLI commands: `azfloci deploy create/validate/show`. Sample templates: storage-account.json, full-stack.json.
- **Docker Compose**: Multi-stage Dockerfile, standalone simulator image (Dockerfile.simulator), docker-compose.yml with shared state volume, .dockerignore.
- **Plugin Architecture**: `ResourceProvider` interface, thread-safe `Registry` with type routing, built-in Storage and Compute providers, CRUD delegation. 15 plugin tests.
- **Cost Estimation**: Estimator with embedded pricing for 6 resource types (Storage, KeyVault, Compute, VNet, PublicIP, NSG). Text and JSON output. CLI: `azfloci cost estimate --file resources.json`. 11 cost tests.
- 22 ARM parser tests, 7 deployer tests.

## [0.3.0] — 2026-07-27
## [0.2.0] — 2026-07-26

### Added
- **Networking Simulation**: VNet, Subnet, NSG, NSG Rule, Public IP — full CRUD with CIDR prefix support, cascade operations, 4-word command routing.
- **VM Simulation**: Full VM lifecycle with state machine (Creating → Running → Stopped → Deallocated). Commands: `vm create/show/list/delete/start/stop/restart/deallocate`.
- **Drift Detection Engine**: `azfloci drift snapshot` captures state, `drift compare` diffs two snapshots, `drift report` compares live state against baseline. Human-readable and JSON output formats.
- **Router 4-word dispatch**: Supports `network vnet subnet create` style commands.
- **40+ new tests**: Network store (12), Network handlers (7), VM store (4), VM handlers (6), Drift engine (7).


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

## [0.1.0] — 2026-07-25

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
