# azfloci

Azure resource orchestration CLI — inspired by [floci](https://github.com/floci-io/floci). Wraps the Azure CLI (`az`) to provide workflow-driven resource management, pipeline execution, and configuration management.

## Architecture

```
┌─────────────────────────────────────────────┐
│                azfloci CLI                  │
│  (cobra commands: auth, group, resource,    │
│   workflow, config)                         │
├──────────┬──────────┬───────────┬───────────┤
│ Azure    │ Resource │ Workflow  │ Config    │
│ Executor │ Manager  │ Engine    │ Manager   │
├──────────┴──────────┴───────────┴───────────┤
│            Azure CLI (`az`)                 │
└─────────────────────────────────────────────┘
```

## Prerequisites
- Go 1.22+
- Azure CLI (`az`) installed and on PATH
- Azure subscription (`az login`)

## Install
```bash
go install github.com/ankurCES/floc-zure/cmd/azfloci@latest
# or
make build && ./bin/azfloci
```

## Quick Start
```bash
azfloci auth status          # verify Azure login
azfloci group create myRG eastus
azfloci resource list -g myRG
azfloci workflow run deploy.yaml
azfloci config set defaults.location westus2
```

## Command Reference
| Command | Description | Docs |
|---|---|---|
| `azfloci version` | Print version | [version](docs/commands/version.md) |
| `azfloci auth status` | Show current Azure account | [auth-status](docs/commands/auth-status.md) |
| `azfloci group create/list/delete` | Resource group CRUD | [group](docs/commands/group.md) |
| `azfloci resource list/show/delete/tag` | Resource operations | [resource](docs/commands/resource.md) |
| `azfloci workflow run/validate` | Execute YAML workflows | [workflow](docs/commands/workflow.md) |
| `azfloci config set/get/list/init` | Manage configuration | [config](docs/commands/config.md) |

## Documentation
- **[Architecture](docs/architecture.md)** — component diagram, data flow, error handling, extension points
- **[Getting Started](docs/guides/getting-started.md)** — installation to first workflow
- **[Azure Setup](docs/guides/azure-setup.md)** — permissions, service principals, CI/CD
- **[Migration from floci](docs/guides/migration-from-floci.md)** — differences and mapping
- **[Development Guide](docs/development.md)** — dev setup, tests, adding commands, conventions
- **[Changelog](CHANGELOG.md)** — release history

## Development
```bash
make test        # unit tests
make e2e-test    # end-to-end (needs az login)
make lint        # golangci-lint
make build       # build binary
```

## License
Apache 2.0 — see [LICENSE](LICENSE)
