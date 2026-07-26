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
| Command | Description |
|---|---|
| `azfloci version` | Print version |
| `azfloci auth status` | Show current Azure account |
| `azfloci group create/list/delete` | Resource group CRUD |
| `azfloci resource list/show/delete/tag` | Resource operations |
| `azfloci workflow run/validate` | Execute YAML workflows |
| `azfloci config set/get/list/init` | Manage configuration |

## Development
```bash
make test        # unit tests
make e2e-test    # end-to-end (needs az login)
make lint        # golangci-lint
make build       # build binary
```

## License
Apache 2.0 — see [LICENSE](LICENSE)
