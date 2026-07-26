# Changelog

All notable changes to azfloci are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] — 2024-01-01

### Added
- **CLI framework**: Cobra-based command tree with `version`, `auth status` commands.
- **Azure CLI executor**: `CLIExecutor` interface wrapping `az` via `os/exec`. Supports `Run()`, `RunJSON()`, `GetAccount()`, `IsAuthenticated()`, `SetSubscription()`.
- **Configuration management**: Viper-based `config.Manager` with YAML file, env vars (`AZFLOCI_*`), and defaults.
- **Resource management interfaces**: `resource.Manager` for resource group CRUD and generic resource operations.
- **Workflow engine interfaces**: `workflow.Engine` for YAML pipeline execution with DAG dependencies, error policies (`continue`/`abort`/`retry`), and concurrent step execution.
- **Data models**: `AzureAccount`, `ResourceGroup`, `Resource`, `Workflow`, `WorkflowStep`, `StepResult`, `WorkflowResult`, `Config`.
- **Error types**: `CLIError`, `AuthError`, `ValidationError`, `WorkflowError` (with `Unwrap()`).
- **E2E test suite**: 20+ tests covering auth, config, resources, workflows, error handling, and version output. Test helpers: `CLIRunner`, `SetupResourceGroup`, assertion utilities.
- **Build system**: Makefile with `build`, `test`, `e2e-test`, `lint`, `fmt`, `vet` targets.
- **GoReleaser config**: Cross-compilation for linux/darwin/windows × amd64/arm64.
- **Documentation**: Architecture doc, command reference, getting started guide, Azure setup guide, migration guide, development guide.
