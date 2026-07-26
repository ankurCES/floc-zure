# azfloci group

Manage Azure resource groups.

## Subcommands

| Command | Description |
|---|---|
| `azfloci group create <name> <location>` | Create a resource group |
| `azfloci group list` | List all resource groups |
| `azfloci group delete <name>` | Delete a resource group |

## Examples
```bash
azfloci group create myRG eastus
azfloci group list
azfloci group delete myRG
```

## Flags
- `--tags key=value` — Tags to apply on create

## Related Commands
- `azfloci resource list -g myRG`
