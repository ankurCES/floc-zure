# azfloci Roadmap

## Phase 1: CLI Framework + Azure CLI Integration + Auth ✅
- Cobra CLI skeleton (`azfloci version`, `azfloci auth status`)
- Azure CLI executor (shell out to `az`, parse JSON)
- Auth verification (`az account show`)
- Config loading (viper)

## Phase 2: Resource Management Commands ✅
- `azfloci group create/list/show/delete` — resource groups
- `azfloci resource list/show/delete/tag` — generic resources

## Phase 3: Workflow/Pipeline Engine ✅
- YAML workflow definitions with step DAGs
- Topological sort, concurrent independent steps
- Error policies: continue / abort / retry
- `azfloci workflow run/validate/list`

## Phase 4: Configuration Management ✅
- `azfloci config set/get/list/init`
- Config file (~/.azfloci.yaml), env vars, CLI flags
- Default resource group, location, tags

## Phase 5: Azure Cloud Simulator ✅
- Drop-in fake `az` CLI binary (`az-simulator`)
- JSON-backed state store with auto-seed (test subscription, no real Azure needed)
- Account, resource group, and resource CRUD handlers
- `AZFLOCI_AZ_PATH` env var integration

## Phase 6: Storage Account Simulation ✅
- `az storage account create/show/list/delete`
- `az storage container create/show/list/delete`
- `az storage blob upload/show/list/delete`
- Cascade deletes (account → containers → blobs)

## Phase 7: Key Vault Simulation ✅
- `az keyvault create/show/list/delete`
- `az keyvault secret set/show/list/delete` (with versioning)
- `az keyvault key create/show/list/delete`
- Cascade deletes (vault → secrets + keys)

## Phase 8: CI/CD Pipeline ✅
- GitHub Actions: matrix build (Go 1.22/1.23 × ubuntu/macos)
- Lint (golangci-lint), unit tests, build, simulator E2E tests
- Makefile targets: `lint-ci`, `test-all`, `test-sim-unit`, `sim-test`

## Phase 9: Networking Simulation ✅
- `az network vnet create/show/list/delete`
- `az network nsg create/show/list/delete`, `nsg rule add`
- `az network public-ip create/show/list/delete`
- CIDR validation, subnet carving

## Phase 10: VM Simulation ✅
- `az vm create/show/list/delete/start/stop/restart/deallocate`
- State machine (Creating → Running → Stopped → Deallocated)

## Phase 11: Drift Detection ✅
- `azfloci drift snapshot` — capture current state
- `azfloci drift compare` — diff live vs. snapshot
- `azfloci drift report` — human-readable + JSON diff

## Phase 12: ARM/Bicep Template Engine 🔜
- `azfloci deploy create --template template.json --parameters params.json`
- Parse ARM templates → create simulated resources

## Phase 13: Docker Compose Environment 🔜
- Full `docker-compose.yml` with simulator + Azurite
- One-command dev setup: `docker compose up`
