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
  "resources": {}
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
