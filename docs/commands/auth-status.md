# azfloci auth status

Show current Azure authentication state.

## Synopsis
```
azfloci auth status
```

## Description
Calls `az account show` under the hood. Displays subscription name/ID, tenant ID, and user info. Exits non-zero with guidance if not authenticated.

## Examples
```bash
azfloci auth status
# Authenticated ✓
#   Subscription: My Sub (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
#   Tenant:       yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
#   User:         user@example.com (user)
```

## Related Commands
- `az login` — authenticate before using azfloci
- `azfloci config set subscription <id>` — switch subscription
