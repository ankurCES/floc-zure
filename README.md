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
  <img src="https://img.shields.io/badge/storage-simulated-blueviolet?style=flat-square&logo=microsoftazure" alt="Storage Simulated">
  <img src="https://img.shields.io/badge/keyvault-simulated-blueviolet?style=flat-square&logo=microsoftazure" alt="Key Vault Simulated">
</p>

---

## 🚀 What is floc-zure?

**floc-zure** is an Azure resource orchestration CLI inspired by [floci](https://github.com/floci-io/floci). It wraps the Azure CLI (`az`) with workflow-driven resource management — and ships with a **built-in Azure cloud simulator** so you can develop and test everything locally without an Azure subscription.

### ✨ Key Features

| Feature | Description |
|---|---|
| ☁️ **Azure Cloud Simulator** | Drop-in fake `az` CLI — runs entirely offline with a local JSON state file |
| 💾 **Storage Account Simulation** | Full CRUD for storage accounts, blob containers, and blobs |
| 🔐 **Key Vault Simulation** | Vaults, secrets (with versioning), and cryptographic keys |
| 🔁 **Workflow Engine** | YAML-defined pipelines with DAG dependency resolution, concurrent execution, retry & error policies |
| 📦 **Resource Management** | Create, list, show, delete, and tag resource groups and resources |
| ⚙️ **Config Management** | Persistent config with defaults for location, subscription, output format |
| 🧪 **No Subscription Needed** | Simulator auto-seeds a test subscription — zero Azure setup required |
| 🏗️ **CI Pipeline** | GitHub Actions matrix build (Go 1.22/1.23 × ubuntu/macos), lint, test, sim-test |

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
azfloci group create --name my-rg --location westus2
azfloci resource list
azfloci workflow run examples/sample_workflow.yaml
azfloci config set defaults.location eastus
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  azfloci CLI (Cobra)                                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ auth     │ │ group    │ │ workflow │ │ config   │           │
│  │ status   │ │ create   │ │ run      │ │ set/get  │           │
│  └────┬─────┘ │ list     │ │ validate │ │ list     │           │
│       │       │ show     │ │ list     │ │ init     │           │
│       │       │ delete   │ └────┬─────┘ └──────────┘           │
│       │       └────┬─────┘      │                               │
│       ▼            ▼            ▼                               │
│  ┌──────────────────────────────────────┐                       │
│  │  CLIExecutor (internal/azure)        │                       │
│  │  ┌─────────────────────────────────┐ │                       │
│  │  │ AZFLOCI_AZ_PATH env var         │ │                       │
│  │  │  → real `az` CLI               │ │                       │
│  │  │  → or simulator binary          │ │                       │
│  │  └─────────────────────────────────┘ │                       │
│  └──────────────────┬───────────────────┘                       │
│                     │                                           │
├─────────────────────┼───────────────────────────────────────────┤
│  Azure Cloud Sim    │                                           │
│  ┌──────────────────▼───────────────────┐                       │
│  │  Router → Handlers                   │                       │
│  │  ┌──────────┐ ┌──────────────────┐   │                       │
│  │  │ account  │ │ storage account  │   │                       │
│  │  │ group    │ │ storage container│   │                       │
│  │  │ resource │ │ storage blob     │   │                       │
│  │  │          │ │ keyvault         │   │                       │
│  │  │          │ │ keyvault secret  │   │                       │
│  │  │          │ │ keyvault key     │   │                       │
│  │  └──────────┘ └──────────────────┘   │                       │
│  │              │                       │                       │
│  │  ┌───────────▼───────────────────┐   │                       │
│  │  │  State Store (JSON file)      │   │                       │
│  │  │  ~/.azfloci-sim/state.json    │   │                       │
│  │  └───────────────────────────────┘   │                       │
│  └──────────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Simulator Commands

The Azure Cloud Simulator implements the following `az` commands locally:

### Account Management

| Command | Description |
|---|---|
| `az account show` | Show active subscription |
| `az account set --subscription ID` | Switch subscription |
| `az account list` | List all subscriptions |

### Resource Groups

| Command | Description |
|---|---|
| `az group create -n NAME -l LOCATION` | Create a resource group |
| `az group show -n NAME` | Show a resource group |
| `az group list` | List all resource groups |
| `az group delete -n NAME --yes` | Delete a resource group (cascades) |

### Resources

| Command | Description |
|---|---|
| `az resource list [-g RG]` | List resources (optionally by group) |
| `az resource show --ids ID` | Show a resource by ARM ID |
| `az resource delete --ids ID --yes` | Delete a resource |
| `az resource tag --ids ID --tags k=v` | Merge tags onto a resource |

### 💾 Storage Accounts, Containers & Blobs

| Command | Description |
|---|---|
| `az storage account create -n NAME -g RG -l LOC` | Create storage account |
| `az storage account show -n NAME` | Show storage account |
| `az storage account list [-g RG]` | List storage accounts |
| `az storage account delete -n NAME` | Delete (cascades containers & blobs) |
| `az storage container create -n NAME --account-name ACCT` | Create blob container |
| `az storage container show -n NAME --account-name ACCT` | Show container |
| `az storage container list --account-name ACCT` | List containers |
| `az storage container delete -n NAME --account-name ACCT` | Delete container (cascades blobs) |
| `az storage blob upload -n NAME --account-name A -c C` | Upload a blob |
| `az storage blob show -n NAME --account-name A -c C` | Show blob metadata |
| `az storage blob list --account-name A -c C` | List blobs in container |
| `az storage blob delete -n NAME --account-name A -c C` | Delete a blob |

### 🔐 Key Vault, Secrets & Keys

| Command | Description |
|---|---|
| `az keyvault create -n NAME -g RG -l LOC` | Create a key vault |
| `az keyvault show -n NAME` | Show vault details |
| `az keyvault list [-g RG]` | List vaults |
| `az keyvault delete -n NAME` | Delete vault (cascades secrets & keys) |
| `az keyvault secret set -n NAME --vault-name V --value VAL` | Set a secret (versioned) |
| `az keyvault secret show -n NAME --vault-name V` | Show latest secret version |
| `az keyvault secret list --vault-name V` | List secrets |
| `az keyvault secret delete -n NAME --vault-name V` | Delete a secret |
| `az keyvault key create -n NAME --vault-name V [--kty RSA]` | Create a key |
| `az keyvault key show -n NAME --vault-name V` | Show key details |
| `az keyvault key list --vault-name V` | List keys |
| `az keyvault key delete -n NAME --vault-name V` | Delete a key |

---

## 🔁 Workflow Engine

Define multi-step Azure operations as YAML workflows with dependency DAGs:

```yaml
name: deploy-app
steps:
  - name: create-rg
    command: az group create -n app-rg -l eastus
    error_policy: abort

  - name: create-storage
    command: az storage account create -n appstore -g app-rg -l eastus
    depends_on: [create-rg]
    error_policy: retry
    max_retries: 3

  - name: create-vault
    command: az keyvault create -n app-vault -g app-rg -l eastus
    depends_on: [create-rg]
    error_policy: abort

  - name: set-connection-string
    command: az keyvault secret set -n conn-str --vault-name app-vault --value "DefaultEndpoint=..."
    depends_on: [create-vault, create-storage]
```

```bash
azfloci workflow run deploy.yaml    # execute
azfloci workflow validate deploy.yaml  # validate DAG (no execution)
azfloci workflow list               # list available workflows
```

**Features:**
- Topological sort ensures correct execution order
- Independent steps run concurrently
- Error policies: `continue`, `abort`, `retry` (with configurable max retries)
- Full result reporting per step

---

## ⚙️ Configuration

```bash
azfloci config init                    # create ~/.azfloci.yaml with defaults
azfloci config set defaults.location westus2
azfloci config set defaults.resource_group my-rg
azfloci config get defaults.location   # → westus2
azfloci config list                    # show all settings
```

Environment variables override config file values:
```bash
AZFLOCI_SUBSCRIPTION_ID=xxx
AZFLOCI_DEFAULTS_LOCATION=eastus
AZFLOCI_OUTPUT_FORMAT=table
AZFLOCI_AZ_PATH=/path/to/az-simulator   # use simulator instead of real az
```

---

## 🧪 Testing

```bash
make test              # unit tests (all packages)
make test-sim-unit     # simulator unit tests only
make test-all          # unit + simulator tests
make sim-test          # full e2e with simulator binary
make lint              # golangci-lint
make vet               # go vet
```

### Test matrix (CI)

| OS | Go 1.22 | Go 1.23 |
|---|---|---|
| Ubuntu | ✅ | ✅ |
| macOS | ✅ | ✅ |

---

## 📁 Project Structure

```
floc-zure/
├── cmd/azfloci/              # CLI entry point
├── internal/
│   ├── azure/                # CLIExecutor (shells out to az / simulator)
│   ├── cli/                  # Cobra command definitions
│   ├── config/               # Viper-based config manager
│   ├── resource/             # Resource group & resource CRUD
│   └── workflow/             # YAML workflow engine + DAG resolver
├── pkg/
│   ├── models/               # Shared types (Account, Resource, Workflow, etc.)
│   └── utils/                # Error types, helpers
├── simulator/
│   ├── cmd/az/               # Simulator binary (fake az CLI)
│   └── internal/
│       ├── handlers/         # Command handlers (account, group, storage, keyvault)
│       ├── router/           # Arg-based command router
│       └── state/            # JSON-backed state store
├── configs/                  # Example config files
├── docs/                     # Architecture, guides, command reference
├── tests/e2e/                # End-to-end test suite
├── .github/workflows/ci.yml  # CI pipeline
├── Makefile                  # Build, test, lint targets
└── .goreleaser.yml           # Cross-platform release config
```

---

## 🗺️ Roadmap

- [x] **Phase 1** — CLI framework, Azure CLI executor, auth
- [x] **Phase 2** — Resource management (groups + resources)
- [x] **Phase 3** — Workflow engine (YAML DAG, concurrency, retry)
- [x] **Phase 4** — Configuration management
- [x] **Phase 5** — Azure Cloud Simulator (offline testing)
- [x] **Phase 6** — Storage Account simulation (accounts, containers, blobs)
- [x] **Phase 7** — Key Vault simulation (vaults, secrets, keys)
- [x] **Phase 8** — CI/CD pipeline (GitHub Actions matrix)
- [ ] **Phase 9** — Networking simulation (VNet, NSG, public IP)
- [ ] **Phase 10** — VM simulation (with state machine lifecycle)
- [ ] **Phase 11** — Drift detection (snapshot → compare → report)
- [ ] **Phase 12** — ARM/Bicep template deployment
- [ ] **Phase 13** — Docker Compose dev environment

---

## 📚 Documentation

| Document | Description |
|---|---|
| [Architecture](docs/architecture.md) | Component design, data flow, extension points |
| [Simulator Design](docs/simulator-design.md) | Simulator internals, state schema, integration |
| [Getting Started](docs/guides/getting-started.md) | Install → auth → first workflow → cleanup |
| [Azure Setup](docs/guides/azure-setup.md) | Permissions, service principal, GitHub Actions |
| [Migration from floci](docs/guides/migration-from-floci.md) | AWS→Azure concept mapping |
| [Development Guide](docs/development.md) | Contributing, adding commands, test patterns |
| [Command Reference](docs/commands/) | Man-page-style docs for every command |

---

## 📄 License

Apache License 2.0 — see [LICENSE](LICENSE).
