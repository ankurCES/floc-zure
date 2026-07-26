# azfloci resource

Manage Azure resources.

## Subcommands

| Command | Description |
|---|---|
| `azfloci resource list -g <group>` | List resources in a group |
| `azfloci resource show <id>` | Show resource details by ARM ID |
| `azfloci resource delete <id>` | Delete a resource |
| `azfloci resource tag <id> --tags k=v` | Add/update tags |

## Examples
```bash
azfloci resource list -g myRG
azfloci resource show /subscriptions/.../myVM
azfloci resource tag /subscriptions/.../myVM --tags env=prod
```

## Related Commands
- `azfloci group list`
