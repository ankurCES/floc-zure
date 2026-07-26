# azfloci config

Manage azfloci configuration.

## Subcommands

| Command | Description |
|---|---|
| `azfloci config set <key> <value>` | Set a config value |
| `azfloci config get <key>` | Get a config value |
| `azfloci config list` | Show all config |
| `azfloci config init` | Create default config file |

## Config File
Location: `~/.azfloci.yaml` (or `.azfloci.yaml` in CWD).

```yaml
subscription: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
location: "eastus"
output_format: "json"
verbose: false
tags:
  team: "platform"
defaults:
  resource_group: "default-rg"
  location: "eastus"
```

## Environment Variables
All keys can be set via `AZFLOCI_` prefix: `AZFLOCI_LOCATION=westus2`.

## Related Commands
- `azfloci auth status`
