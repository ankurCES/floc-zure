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

### Networking (VNet, Subnet, NSG, Public IP)

| az command | Handler | Notes |
|---|---|---|
| `network vnet create -n N -g RG -l LOC --address-prefixes CIDR` | network.VNetCreate | Persists VNet with address space |
| `network vnet show -n N -g RG` | network.VNetShow | Lookup by name+RG |
| `network vnet list [-g RG]` | network.VNetList | Filter by RG |
| `network vnet delete -n N -g RG` | network.VNetDelete | Cascades subnets |
| `network vnet subnet create -n N --vnet-name V -g RG --address-prefixes CIDR` | network.SubnetCreate | 4-word route |
| `network vnet subnet show -n N --vnet-name V -g RG` | network.SubnetShow | |
| `network vnet subnet list --vnet-name V -g RG` | network.SubnetList | |
| `network vnet subnet delete -n N --vnet-name V -g RG` | network.SubnetDelete | |
| `network nsg create -n N -g RG -l LOC` | network.NSGCreate | |
| `network nsg show -n N -g RG` | network.NSGShow | |
| `network nsg list [-g RG]` | network.NSGList | Filter by RG |
| `network nsg delete -n N -g RG` | network.NSGDelete | Cascades rules |
| `network nsg rule create -n N --nsg-name NSG -g RG --priority P --access A --protocol P --direction D` | network.NSGRuleCreate | 4-word route |
| `network nsg rule delete -n N --nsg-name NSG -g RG` | network.NSGRuleDelete | |
| `network public-ip create -n N -g RG -l LOC` | network.PublicIPCreate | Auto-assigns IP |
| `network public-ip show -n N -g RG` | network.PublicIPShow | |
| `network public-ip list [-g RG]` | network.PublicIPList | Filter by RG |
| `network public-ip delete -n N -g RG` | network.PublicIPDelete | |

### Virtual Machines

| az command | Handler | Notes |
|---|---|---|
| `vm create -n N -g RG -l LOC --image IMG --size SZ` | vm.Create | Initial state: Running |
| `vm show -n N -g RG` | vm.Show | Includes powerState |
| `vm list [-g RG]` | vm.List | Filter by RG |
| `vm delete -n N -g RG --yes` | vm.Delete | |
| `vm start -n N -g RG` | vm.Start | → Running |
| `vm stop -n N -g RG` | vm.Stop | → Stopped |
| `vm restart -n N -g RG` | vm.Restart | → Running |
| `vm deallocate -n N -g RG` | vm.Deallocate | → Deallocated |

**VM State Machine:**
```
Creating → Running ↔ Stopped → Deallocated
                  ↔ Deallocated
       restart ──→ Running
```

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
  "keys": {},
  "vnets": {},
  "subnets": {},
  "nsgs": {},
  "nsg_rules": {},
  "public_ips": {},
  "vms": {},
  "servicebus_namespaces": {},
  "servicebus_queues": {},
  "servicebus_topics": {},
  "servicebus_subscriptions": {},
  "servicebus_messages": {},
  "function_apps": {},
  "functions": {},
  "function_invocations": {}
}
```

### Service Bus (Namespace, Queue, Topic, Subscription, Message)

| az command | Handler | Notes |
|---|---|---|
| `servicebus namespace create -n N -g RG -l LOC [--sku Basic]` | servicebus.NamespaceCreate | SKU flag (Basic/Standard/Premium) |
| `servicebus namespace show -n N -g RG` | servicebus.NamespaceShow | |
| `servicebus namespace list [-g RG]` | servicebus.NamespaceList | Filter by RG |
| `servicebus namespace delete -n N -g RG` | servicebus.NamespaceDelete | Cascade deletes queues/topics/subs/messages |
| `servicebus queue create -n N --namespace-name NS -g RG` | servicebus.QueueCreate | |
| `servicebus queue show -n N --namespace-name NS -g RG` | servicebus.QueueShow | |
| `servicebus queue list --namespace-name NS -g RG` | servicebus.QueueList | |
| `servicebus queue delete -n N --namespace-name NS -g RG` | servicebus.QueueDelete | |
| `servicebus topic create -n N --namespace-name NS -g RG` | servicebus.TopicCreate | |
| `servicebus topic show -n N --namespace-name NS -g RG` | servicebus.TopicShow | |
| `servicebus topic list --namespace-name NS -g RG` | servicebus.TopicList | |
| `servicebus topic delete -n N --namespace-name NS -g RG` | servicebus.TopicDelete | Cascade deletes subscriptions |
| `servicebus topic subscription create -n N --topic-name T --namespace-name NS -g RG` | servicebus.SubscriptionCreate | 5-word route |
| `servicebus topic subscription show -n N --topic-name T --namespace-name NS -g RG` | servicebus.SubscriptionShow | |
| `servicebus topic subscription list --topic-name T --namespace-name NS -g RG` | servicebus.SubscriptionList | |
| `servicebus topic subscription delete -n N --topic-name T --namespace-name NS -g RG` | servicebus.SubscriptionDelete | |
| `servicebus queue message send --namespace-name NS --queue-name Q -g RG --body MSG` | servicebus.MessageSend | 5-word route |
| `servicebus queue message receive --namespace-name NS --queue-name Q -g RG` | servicebus.MessageReceive | Dequeues (destructive read) |
| `servicebus queue message peek --namespace-name NS --queue-name Q -g RG` | servicebus.MessagePeek | Non-destructive read |

### Function Apps (App, Function, Invoke)

| az command | Handler | Notes |
|---|---|---|
| `functionapp create -n N -g RG -l LOC --runtime RUNTIME` | functionapp.Create | Runtime + version flags |
| `functionapp show -n N -g RG` | functionapp.Show | |
| `functionapp list [-g RG]` | functionapp.List | Filter by RG |
| `functionapp delete -n N -g RG` | functionapp.Delete | Cascade deletes functions + invocations |
| `functionapp function create --function-app-name APP -g RG -n N --trigger-type TYPE` | functionapp.FunctionCreate | Bindings via trigger-type |
| `functionapp function show --function-app-name APP -g RG -n N` | functionapp.FunctionShow | |
| `functionapp function list --function-app-name APP -g RG` | functionapp.FunctionList | |
| `functionapp function delete --function-app-name APP -g RG -n N` | functionapp.FunctionDelete | |
| `functionapp function invoke --function-app-name APP -g RG -n N [--body JSON]` | functionapp.FunctionInvoke | Simulated echo response |
| `functionapp function invocations --function-app-name APP -g RG -n N` | functionapp.FunctionInvocations | Returns invocation history |

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
