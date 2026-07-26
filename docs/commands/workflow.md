# azfloci workflow

Execute YAML-defined Azure provisioning workflows.

## Subcommands

| Command | Description |
|---|---|
| `azfloci workflow run <file>` | Execute a workflow |
| `azfloci workflow validate <file>` | Validate without executing |

## Workflow YAML Format
```yaml
name: "deploy"
steps:
  - name: "create-rg"
    command: "group create"
    args: { name: "myRG", location: "eastus" }
  - name: "verify"
    command: "group show"
    args: { name: "myRG" }
    depends_on: ["create-rg"]
    on_error: "retry"
    max_retries: 3
```

## Step Fields
| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique step identifier |
| `command` | yes | Azure CLI command to run |
| `args` | no | Key-value arguments |
| `depends_on` | no | Steps that must complete first |
| `on_error` | no | `continue`, `abort`, or `retry` |
| `max_retries` | no | Retry count (with `on_error: retry`) |

## Related Commands
- `azfloci config set defaults.location westus2`
