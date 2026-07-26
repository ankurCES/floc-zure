# Azure Setup Guide

## Required Permissions
azfloci needs an identity with:
- `Microsoft.Resources/subscriptions/resourceGroups/*` — resource group CRUD
- `Microsoft.Resources/subscriptions/resources/read` — list resources
- `Microsoft.Resources/subscriptions/tagNames/*` — tagging

The built-in **Contributor** role covers all of these.

## Interactive Login (Development)
```bash
az login
az account set --subscription "YOUR_SUBSCRIPTION_ID"
azfloci auth status
```

## Service Principal (CI/CD)
```bash
# Create service principal with Contributor role
az ad sp create-for-rbac \
  --name "azfloci-ci" \
  --role Contributor \
  --scopes /subscriptions/YOUR_SUBSCRIPTION_ID

# Output:
# {
#   "appId": "...",
#   "password": "...",
#   "tenant": "..."
# }

# Login as service principal
az login --service-principal \
  -u APP_ID \
  -p PASSWORD \
  --tenant TENANT_ID
```

## GitHub Actions Example
```yaml
- name: Azure Login
  uses: azure/login@v2
  with:
    creds: ${{ secrets.AZURE_CREDENTIALS }}

- name: Run azfloci workflow
  run: |
    azfloci auth status
    azfloci workflow run deploy.yaml
```

## Environment Variables
| Variable | Description |
|---|---|
| `AZFLOCI_LOCATION` | Default Azure region |
| `AZFLOCI_VERBOSE` | Enable verbose output |
| `AZURE_SUBSCRIPTION_ID` | Used by `az` CLI |
