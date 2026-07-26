# Migrating from floci

## What is floci?
[floci](https://github.com/floci-io/floci) is a local AWS emulator (Java/Quarkus) that emulates AWS services on a single port. It lets you run AWS SDK/CLI commands against a local endpoint.

## How azfloci Differs

| Aspect | floci | azfloci |
|---|---|---|
| **Cloud** | AWS | Azure |
| **Language** | Java 25 / Quarkus | Go |
| **Approach** | Local emulator (fake services) | CLI wrapper (real Azure) |
| **Auth** | Fake credentials | Real `az login` |
| **Network** | Local HTTP endpoint | Shells out to `az` CLI |
| **Resources** | Emulated in-memory | Real Azure resources |
| **Cost** | Free (local) | Azure subscription costs |

## Key Concept Mappings

| floci (AWS) | azfloci (Azure) |
|---|---|
| `awslocal` wrapper | `azfloci` CLI |
| CloudFormation stacks | Workflow YAML pipelines |
| AWS regions | Azure locations |
| AWS accounts | Azure subscriptions |
| IAM roles/policies | Azure RBAC / Entra ID |
| S3 buckets | Azure Blob Storage |
| SQS queues | Azure Service Bus Queues |

## Migration Steps
1. Install azfloci and Azure CLI
2. Authenticate: `az login` (replaces fake AWS credentials)
3. Translate AWS resource commands to Azure equivalents
4. Convert CloudFormation/Terraform to azfloci workflow YAML
5. Run `azfloci workflow validate` before executing

## What's NOT Ported
- Local emulation — azfloci works against real Azure
- Docker-based service faking
- Multi-service protocol multiplexing
