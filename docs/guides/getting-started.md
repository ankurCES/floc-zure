# Getting Started with azfloci

## Prerequisites
1. **Go 1.22+** — [go.dev/dl](https://go.dev/dl/)
2. **Azure CLI** — [Install guide](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
3. **Azure subscription** — [Free account](https://azure.microsoft.com/free/)

## Install
```bash
# From source
go install github.com/ankurCES/floc-zure/cmd/azfloci@latest

# Or clone & build
git clone https://github.com/ankurCES/floc-zure.git
cd floc-zure
make build
./bin/azfloci version
```

## Authenticate
```bash
az login
azfloci auth status   # verify
```

## First Workflow
1. Create a workflow file `deploy.yaml`:
```yaml
name: "first-deploy"
steps:
  - name: "create-rg"
    command: "group create"
    args:
      name: "azfloci-demo"
      location: "eastus"
  - name: "verify"
    command: "group show"
    args:
      name: "azfloci-demo"
    depends_on: ["create-rg"]
```

2. Validate and run:
```bash
azfloci workflow validate deploy.yaml
azfloci workflow run deploy.yaml
```

3. Clean up:
```bash
azfloci group delete azfloci-demo
```

## Next Steps
- [Azure Setup Guide](azure-setup.md) — permissions & CI/CD service principals
- [Command Reference](../commands/) — all CLI commands
- [Architecture](../architecture.md) — how it works under the hood
