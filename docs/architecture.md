# azfloci Architecture

## Executive Summary

azfloci is a Go CLI that wraps the Azure CLI (`az`) to provide workflow-driven resource management. It uses Cobra for command parsing, Viper for configuration, and shells out to `az` for all Azure operations — making it a thin orchestration layer rather than an SDK client.

## Component Diagram

```
┌──────────────────────────────────────────────────────┐
│                    azfloci CLI                       │
│                                                      │
│  cmd/azfloci/main.go → cli.Execute()                │
│                                                      │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │
│  │  auth   │ │ resource │ │ workflow │ │ config  │ │
│  │  cmd    │ │  manager │ │  engine  │ │ manager │ │
│  └────┬────┘ └─────┬────┘ └─────┬────┘ └────┬────┘ │
│       │            │            │            │      │
│  ┌────▼────────────▼────────────▼────────────▼────┐ │
│  │           CLIExecutor (interface)              │ │
│  │  Run() · RunJSON() · GetAccount()             │ │
│  └─────────────────────┬─────────────────────────┘ │
└────────────────────────┼─────────────────────────────┘
                         │ os/exec
                    ┌────▼────┐
                    │ az CLI  │
                    └─────────┘
```

## Package Layout

| Package | Role |
|---|---|
| `cmd/azfloci` | Entry point — calls `cli.Execute()` |
| `internal/cli` | Cobra command tree (root, version, auth) |
| `internal/azure` | `CLIExecutor` interface + `CLIExecutorImpl` (shells out to `az`) |
| `internal/config` | `Manager` interface + Viper-backed implementation |
| `internal/resource` | `Manager` interface for resource group & resource CRUD |
| `internal/workflow` | `Engine` interface for YAML pipeline execution |
| `pkg/models` | Shared data types: `AzureAccount`, `ResourceGroup`, `Resource`, `Workflow`, `Config` |
| `pkg/utils` | Error types: `CLIError`, `AuthError`, `ValidationError`, `WorkflowError` |

## Azure CLI Integration

All Azure operations go through `CLIExecutor`:

1. `CLIExecutorImpl.Run()` calls `exec.CommandContext(ctx, "az", args...)`.
2. Captures stdout, stderr, exit code into `CLIResult`.
3. Non-zero exit → `CLIError` with command string and stderr.
4. `RunJSON()` appends `--output json` and unmarshals stdout into the destination struct.

This design means:
- **No Azure SDK dependency** — only the `az` binary.
- **Testable** — swap `CLIExecutor` interface with a mock.
- **Auth delegation** — relies on `az login`; no token management.

## Error Handling Strategy

Four error types in `pkg/utils`:

| Type | When |
|---|---|
| `CLIError` | `az` command returns non-zero exit code |
| `AuthError` | User not authenticated (`az account show` fails) |
| `ValidationError` | Invalid input/config (field + message) |
| `WorkflowError` | Step failure during workflow execution (wraps cause via `Unwrap()`) |

All errors implement the `error` interface. `WorkflowError` supports `errors.Unwrap()` for chain inspection.

## Configuration

Viper loads config in priority order: CLI flags → env vars (`AZFLOCI_*`) → config file → defaults.

Config file: `~/.azfloci.yaml` or `.azfloci.yaml` in CWD.

Defaults: `output_format=json`, `location=eastus`, `verbose=false`.

## Workflow Engine

The `Engine` interface supports:
- **LoadWorkflow**: Parse YAML into `Workflow` struct.
- **Validate**: Check for cycles, missing deps, invalid commands.
- **Execute**: Topological sort of step DAG, concurrent execution of independent steps.
- **Error policies**: `on_error` per step — `continue`, `abort`, or `retry` with `max_retries`.

Workflow YAML example:
```yaml
name: "deploy-app"
steps:
  - name: "create-rg"
    command: "group create"
    args: { name: "myRG", location: "eastus" }
  - name: "verify-rg"
    command: "group show"
    args: { name: "myRG" }
    depends_on: ["create-rg"]
    on_error: "retry"
    max_retries: 3
```

## Extension Points

1. **New commands**: Add a Cobra command in `internal/cli/`, register in `root.go`'s `init()`.
2. **New resource operations**: Extend the `resource.Manager` interface.
3. **Custom workflow steps**: Implement `StepExecutor` interface.
4. **Alternative executors**: Implement `CLIExecutor` (e.g., for Azure SDK-based execution or mocking).
5. **Output formats**: `Config.OutputFormat` supports `json`, `table`, `yaml`.

## Build & Release

- **Build**: `go build` with ldflags injecting version from git tags.
- **Release**: GoReleaser cross-compiles for linux/darwin/windows × amd64/arm64.
- **CI**: `make test` (unit), `make e2e-test` (integration with real Azure).
