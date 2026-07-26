# azfloci Roadmap

## Module 1: CLI Framework + Azure CLI Integration + Auth ✅
- Cobra CLI skeleton (`azfloci version`, `azfloci auth status`)
- Azure CLI executor (shell out to `az`, parse JSON)
- Auth verification (`az account show`)
- Config loading (viper)

## Module 2: Resource Management Commands
- `azfloci group create/list/delete` — resource groups
- `azfloci resource list/show/delete/tag` — generic resources
- Depends on: Module 1

## Module 3: Workflow/Pipeline Engine
- YAML workflow definitions with step DAGs
- Topological sort, concurrent independent steps
- Error policies: continue / abort / retry
- `azfloci workflow run/validate/list`
- Depends on: Modules 1, 2

## Module 4: Configuration Management
- `azfloci config set/get/list/init`
- Config file (~/.azfloci.yaml), env vars, CLI flags
- Default resource group, location, tags
- Depends on: Module 1
