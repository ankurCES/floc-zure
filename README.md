<p align="center">
  <img src="assets/logo.svg" alt="floc-zure logo" width="600"/>
</p>

<p align="center">
  <strong>Azure Cloud Simulator & Orchestration CLI</strong><br/>
  <em>Test Azure workflows locally — no subscription required.</em>
</p>

<p align="center">
  <a href="https://github.com/ankurCES/floc-zure/actions"><img src="https://img.shields.io/github/actions/workflow/status/ankurCES/floc-zure/ci.yml?branch=main&style=flat-square&logo=github&label=CI" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/ankurCES/floc-zure"><img src="https://goreportcard.com/badge/github.com/ankurCES/floc-zure?style=flat-square" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/ankurCES/floc-zure"><img src="https://img.shields.io/badge/godoc-reference-5272B4?style=flat-square&logo=go" alt="GoDoc"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square" alt="License"></a>
  <a href="https://github.com/ankurCES/floc-zure/releases"><img src="https://img.shields.io/github/v/release/ankurCES/floc-zure?style=flat-square&color=orange" alt="Release"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go 1.22+">
  <img src="https://img.shields.io/badge/Azure-Simulator-0078d4?style=flat-square&logo=microsoftazure" alt="Azure Simulator">
  <img src="https://img.shields.io/badge/subscription-not%20required-success?style=flat-square" alt="No Subscription Required">
</p>

---

## 🚀 What is floc-zure?

**floc-zure** is an Azure resource orchestration CLI inspired by [floci](https://github.com/floci-io/floci). It wraps the Azure CLI (`az`) with workflow-driven resource management — and ships with a **built-in Azure cloud simulator** so you can develop and test everything locally without an Azure subscription.

### ✨ Key Features

| Feature | Description |
|---|---|
| ☁️ **Azure Cloud Simulator** | Drop-in fake `az` CLI — runs entirely offline with a local JSON state file |
| 🔁 **Workflow Engine** | YAML-defined pipelines with DAG dependency resolution, concurrent execution, retry & error policies |
| 📦 **Resource Management** | Create, list, show, delete, and tag resource groups and resources |
| ⚙️ **Config Management** | Persistent config with defaults for location, subscription, output format |
| 🧪 **No Subscription Needed** | Simulator auto-seeds a test subscription — zero Azure setup required |
| 🐳 **Docker Support** | Containerized testing with `docker-compose up` |

---

## 📦 Installation

```bash
# From source
go install github.com/ankurCES/floc-zure/cmd/azfloci@latest

# Or clone and build
git clone https://github.com/ankurCES/floc-zure.git
cd floc-zure
make build          # → bin/azfloci
make sim-build      # → bin/az-simulator
```

---

## ⚡ Quick Start (No Azure Subscription Required)

The simulator seeds a test subscription automatically. No `az login` needed.

```bash
# 1. Build both binaries
make build sim-build

# 2. Point azfloci at the simulator
export AZFLOCI_AZ_PATH=./bin/az-simulator

# 3. Use it exactly like real Azure
azfloci auth status
# Authenticated ✓
#   Subscription: Simulated-Subscription-1 (00000000-0000-0000-0000-000000000001)
#   Tenant:       00000000-0000-0000-0000-000000000099
#   User:         simulator@azfloci.local (user)

azfloci group create --name my-rg --location westus2
azfloci resource list
azfloci workflow run examples/sample_workflow.yaml
azfloci config set defaults.location eastus
```

### 🧪 Run All Tests Offline

```bash
make sim-test    # builds simulator, runs full e2e suite — no Azure needed
make test        # unit tests (always offline)
```

---

## ☁️ Azure Cloud Simulator

The simulator is a **standalone Go binary** that replaces the `az` CLI. It maintains state in a local JSON file and responds to the same commands azfloci uses.

### How It Works

```
azfloci CLI ──shells out──► az binary (real or simulator)
                                │
                          ┌─────▼──────┐
                          │  Simulator  │
                          │  Router     │──► account show/set/list
                          │             │──► group create/show/list/delete
                          │             │──► resource list/show/delete/tag
                          └─────┬──────┘
                                │
                          ┌─────▼──────┐
                          │ State Store│
                          │ (JSON file)│
                          └────────────┘
```

### Seeded Test Data

On first run, the simulator creates a state file with:

| Field | Value |
|---|---|
| Subscription ID | `00000000-0000-0000-0000-000000000001` |
| Subscription Name | `Simulated-Subscription-1` |
| Tenant ID | `00000000-0000-0000-0000-000000000099` |
| User | `simulator@azfloci.local` |
| State | `Enabled` |

No real Azure credentials or subscription needed. The state file lives at `~/.azfloci-sim/state.json` (override with `AZFLOCI_SIM_STATE`).

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `AZFLOCI_AZ_PATH` | Path to `az` binary (set to simulator for offline use) | `az` from PATH |
| `AZFLOCI_SIM_STATE` | Path to simulator state JSON file | `~/.azfloci-sim/state.json` |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────┐
│                  azfloci CLI                     │
│  (cobra: auth, group, resource, workflow, config)│
├───────────┬───────────┬───────────┬──────────────┤
│  Azure    │ Resource  │ Workflow  │   Config     │
│  Executor │ Manager   │ Engine    │   Manager    │
├───────────┴───────────┴───────────┴──────────────┤
│         Azure CLI  (real `az` or simulator)      │
└──────────────────────────────────────────────────┘

simulator/
├── cmd/az/main.go              # Fake az binary entry point
├── internal/
│   ├── state/store.go          # JSON-backed state store
│   ├── router/router.go        # Command dispatcher
│   └── handlers/               # account, group, resource handlers
```

---

## 📋 Command Reference

| Command | Description |
|---|---|
| `azfloci version` | Print version |
| `azfloci auth status` | Show current Azure account |
| `azfloci group create --name N --location L` | Create resource group |
| `azfloci group list` | List resource groups |
| `azfloci group show --name N` | Show resource group details |
| `azfloci group delete --name N` | Delete resource group |
| `azfloci resource list [--resource-group RG]` | List resources |
| `azfloci resource show --ids ID` | Show resource details |
| `azfloci resource delete --ids ID` | Delete resource |
| `azfloci resource tag --ids ID --tags k=v` | Tag resource |
| `azfloci workflow run FILE` | Execute YAML workflow |
| `azfloci workflow validate FILE` | Validate workflow syntax |
| `azfloci config init` | Initialize config file |
| `azfloci config set KEY VALUE` | Set config value |
| `azfloci config get KEY` | Get config value |
| `azfloci config list` | Show all config |

---

## 🔁 Workflow Engine

Define infrastructure-as-code pipelines in YAML:

```yaml
name: deploy-app
description: Create RG + storage account
steps:
  - name: create-rg
    command: group create
    args:
      name: my-app-rg
      location: eastus

  - name: create-storage
    command: resource create
    args:
      resource-group: my-app-rg
      name: myappstorage
    depends_on: [create-rg]
    on_error: retry
    max_retries: 3
```

Features: DAG dependency resolution · concurrent execution of independent steps · retry/abort/continue error policies.

---

## 🐳 Docker

```bash
# Run full test suite in container (no local Go or Azure needed)
docker-compose up --build sim-test

# Or build the simulator image
docker build -f simulator/Dockerfile -t az-simulator .
```

---

## 🧑‍💻 Development

```bash
make build       # Build azfloci binary
make sim-build   # Build simulator binary
make test        # Unit tests
make sim-test    # E2E tests with simulator (no Azure needed)
make e2e-test    # E2E tests against real Azure (needs az login)
make lint        # golangci-lint
make fmt         # gofmt
make vet         # go vet
```

---

## 📚 Documentation

| Doc | Description |
|---|---|
| [Architecture](docs/architecture.md) | Component diagram, data flow, extension points |
| [Simulator Design](docs/simulator-design.md) | Simulator architecture and integration |
| [Getting Started](docs/guides/getting-started.md) | Install → first workflow |
| [Azure Setup](docs/guides/azure-setup.md) | Permissions, service principals, CI/CD |
| [Migration from floci](docs/guides/migration-from-floci.md) | Mapping AWS→Azure concepts |
| [Development Guide](docs/development.md) | Dev setup, tests, adding commands |
| [Changelog](CHANGELOG.md) | Release history |

---

## 🗺️ Roadmap

- [x] CLI skeleton + Azure executor
- [x] Resource group & resource management
- [x] Workflow engine with DAG execution
- [x] Config management
- [x] Azure cloud simulator (offline testing)
- [ ] Additional resource providers (VMs, storage accounts, networking)
- [ ] Drift detection
- [ ] State export/import
- [ ] GitHub Actions workflow

---

## 📄 License

Apache 2.0 — see [LICENSE](LICENSE)
