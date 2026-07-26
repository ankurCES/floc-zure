# Azure Cloud Simulator — Design Document

## Overview

The Azure Cloud Simulator is a **fake `az` CLI binary** that mimics Azure Resource Manager
responses without requiring a real Azure subscription. It enables fully offline,
deterministic end-to-end testing of azfloci.

## Architecture

```
┌──────────────────────────────────────────────────┐
│  azfloci CLI                                     │
│  (unchanged — shells out to `az`)                │
│                                                  │
│  CLIExecutorImpl.azPath ──► AZFLOCI_AZ_PATH env  │
│                              │                   │
│                              ▼                   │
│  ┌──────────────────────────────────────────┐    │
│  │  simulator/cmd/az/main.go                │    │
│  │  (drop-in replacement for `az`)          │    │
│  │                                          │    │
│  │  ┌─────────┐  ┌──────────────────────┐   │    │
│  │  │ Router  │─►│ Handlers             │   │    │
│  │  │ (args)  │  │  account show/set    │   │    │
│  │  └─────────┘  │  group CRUD          │   │    │
│  │               │  resource CRUD/tag   │   │    │
│  │               │  storage acct/ctr/bl │   │    │
│  │               │  keyvault/secret/key │   │    │
│  │               └──────────┬───────────┘   │    │
│  │                          │               │    │
│  │               ┌──────────▼───────────┐   │    │
│  │               │ State Store          │   │    │
│  │               │ (JSON file-backed)   │   │    │
│  │               │ ~/.azfloci-sim/      │   │    │
│  │               │   state.json         │   │    │
│  │               └──────────────────────┘   │    │
│  └──────────────────────────────────────────┘    │
└──────────────────────────────────────────────────┘
```

## Commands Simulated

### Account & Resource Management

| az command | Handler | Notes |
|---|---|---|
| `account show` | account.Show | Returns active subscription |
| `account set --subscription ID` | account.Set | Switches active sub |
| `account list` | account.List | Lists all subs |
| `group create --name N --location L [--tags]` | group.Create | Persists to state |
| `group show --name N` | group.Show | Lookup by name |
| `group list` | group.List | All groups |
| `group delete --name N --yes [--no-wait]` | group.Delete | Removes + resources |
| `resource list [--resource-group RG]` | resource.List | Filter by RG |
| `resource show --ids ID` | resource.Show | Lookup by ARM ID |
| `resource delete --ids ID --yes` | resource.Delete | Removes from state |
| `resource tag --ids ID --tags k=v` | resource.Tag | Merges tags |

### Storage Accounts, Containers & Blobs

| az command | Handler | Notes |
|---|---|---|
| `storage account create -n N -g RG -l LOC` | storage.AccountCreate | SKU, kind flags |
| `storage account show -n N` | storage.AccountShow | Lookup by name |
| `storage account list [-g RG]` | storage.AccountList | Filter by RG |
| `storage account delete -n N` | storage.AccountDelete | Cascade deletes containers + blobs |
| `storage container create -n N --account-name A` | storage.ContainerCreate | |
| `storage container show -n N --account-name A` | storage.ContainerShow | |
| `storage container list --account-name A` | storage.ContainerList | |
| `storage container delete -n N --account-name A` | storage.ContainerDelete | Cascade deletes blobs |
| `storage blob upload -n N --account-name A -c C` | storage.BlobUpload | Content-type, source flags |
| `storage blob show -n N --account-name A -c C` | storage.BlobShow | |
| `storage blob list --account-name A -c C` | storage.BlobList | |
| `storage blob delete -n N --account-name A -c C` | storage.BlobDelete | |

### Key Vault, Secrets & Keys

| az command | Handler | Notes |
|---|---|---|
| `keyvault create -n N -g RG -l LOC` | keyvault.Create | SKU flag |
| `keyvault show -n N` | keyvault.Show | |
| `keyvault list [-g RG]` | keyvault.List | Filter by RG |
| `keyvault delete -n N` | keyvault.Delete | Cascade deletes secrets + keys |
| `keyvault secret set -n N --vault-name V --value VAL` | keyvault.SecretSet | Auto-versioned |
| `keyvault secret show -n N --vault-name V` | keyvault.SecretShow | Returns latest version |
| `keyvault secret list --vault-name V` | keyvault.SecretList | |
| `keyvault secret delete -n N --vault-name V` | keyvault.SecretDelete | |
| `keyvault key create -n N --vault-name V [--kty RSA]` | keyvault.KeyCreate | RSA/EC, size flag |
| `keyvault key show -n N --vault-name V` | keyvault.KeyShow | |
| `keyvault key list --vault-name V` | keyvault.KeyList | |
| `keyvault key delete -n N --vault-name V` | keyvault.KeyDelete | |

## State Store

- **Location**: `$AZFLOCI_SIM_STATE` env var, or `~/.azfloci-sim/state.json`
- **Format**: JSON file, read on startup, written after every mutation
- **Thread safety**: sync.RWMutex (the fake az binary is short-lived per invocation, but future HTTP mode may need it)
- **Seed data**: Pre-populated with one subscription and one resource group on first run

### Schema

```json
{
  "active_subscription": "00000000-0000-0000-0000-000000000001",
  "subscriptions": [
    {
      "id": "00000000-0000-0000-0000-000000000001",
      "name": "Simulated-Subscription-1",
      "state": "Enabled",
      "isDefault": true,
      "tenantId": "00000000-0000-0000-0000-000000000099",
      "homeTenantId": "00000000-0000-0000-0000-000000000099",
      "environmentName": "AzureCloud",
      "user": { "name": "simulator@azfloci.local", "type": "user" }
    }
  ],
  "resource_groups": {},
  "resources": {},
  "storage_accounts": {},
  "containers": {},
  "blobs": {},
  "key_vaults": {},
  "secrets": {},
  "keys": {}
}
```

## Integration

1. **Env var**: `AZFLOCI_AZ_PATH` — when set, `CLIExecutorImpl` uses this path instead of `az` from PATH.
2. **E2E tests**: Detect missing real Azure auth → build simulator binary → set `AZFLOCI_AZ_PATH` → run tests offline.
3. **Docker**: `docker-compose.yml` builds both binaries, runs full e2e suite.
4. **Makefile**: `make sim-build`, `make sim-test`.

## Output Format

All handlers output JSON to stdout (matching `az --output json`). Stderr is used
for error messages with the same format as real `az`:
```
ERROR: Resource group 'foo' could not be found.
```

## Exit Codes

- `0` — success
- `1` — general error (resource not found, bad args)
- `2` — usage error (missing required flag)
