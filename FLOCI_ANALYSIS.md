# Floci Codebase Analysis & Azure Mapping Blueprint

## 1. Overview

**Floci** is a free, open-source local AWS emulator (MIT license) built in **Java 25** on **Quarkus 3.36**. It exposes all 68+ AWS services on a single HTTP port (`4566`), letting developers use standard AWS SDKs/CLI/Terraform/CDK against a local Docker container. It is the successor/alternative to LocalStack Community Edition.

**Our goal**: Build **floc-zure** — an analogous tool for **Azure**, written in **Go**, that works with the **Azure CLI** and Azure SDKs.

---

## 2. Architecture

### 2.1 Directory Structure
```
floci/
├── src/main/java/io/github/hectorvent/floci/
│   ├── config/          # EmulatorConfig, TLS, clock
│   ├── core/
│   │   ├── common/      # Request routing, protocols, IAM, Docker, DNS, error handling
│   │   └── storage/     # StorageFactory, InMemory/Persistent/Hybrid/WAL backends
│   ├── lifecycle/       # Startup/shutdown, init hooks, health endpoints
│   └── services/        # 68 service packages (one per AWS service)
│       ├── sqs/         # Controller + Service + model/
│       ├── s3/          # REST XML controller + service + model/
│       ├── lambda/      # Docker-backed execution
│       └── ...
├── src/test/java/       # Unit + Integration tests (JUnit 5, RestAssured)
├── src/main/resources/  # application.yml, pricing snapshots
├── docker/              # Dockerfile, entrypoint, native-image variants
├── compatibility-tests/ # SDK tests: awscli (bats), Go, Java, Node, Python, Terraform, CDK
├── bin/awslocal          # Shell wrapper for AWS CLI → Floci endpoint
├── tools/docs/          # Auto-gen action tables from handler source
├── .github/workflows/   # CI, release, compatibility, docs
├── pom.xml              # Maven build (Quarkus)
└── docker-compose.yml
```

### 2.2 Design Patterns
| Pattern | Floci Implementation | floc-zure Go Equivalent |
|---|---|---|
| Single-port multiplexer | All services on `:4566`, routed by AWS headers (`X-Amz-Target`, `Action`, REST path) | Single `net/http` server with middleware router |
| Service catalog/registry | `ServiceCatalog` + `ServiceDescriptor` records | Go interface `ServiceDescriptor` + registry map |
| Protocol dispatchers | `AwsQueryController`, `AwsJson11Controller`, `AwsJsonCborController` (JAX-RS) | Go HTTP handlers per Azure REST API version |
| Storage abstraction | `StorageFactory` → InMemory / Persistent(JSON) / Hybrid / WAL | Go `StorageBackend` interface with same 4 modes |
| Account/region isolation | `AccountContextFilter`, `RegionResolver` | Middleware extracting subscription/tenant from Azure auth headers |
| Docker integration | `docker-java` for Lambda, RDS, ElastiCache, Neptune containers | Go `docker` client SDK |
| Init hooks | Shell scripts at boot/start/ready/shutdown phases | Same lifecycle hooks |
| TLS | Self-signed cert generation, first-byte HTTP/HTTPS sniffing proxy | Go `crypto/tls` with auto-cert |

### 2.3 Core Abstractions
- **`EmulatorConfig`** — Quarkus `@ConfigMapping` interface tree: port, base-url, hostname, region, account, storage, DNS, auth, security, services (per-service toggles), docker, TLS, init-hooks.
- **`ServiceDescriptor`** — Record: externalKey, configKey, enabled, storageKey, storageMode, protocols, targetPrefixes, credentialScopes.
- **`ServiceCatalog`** — Immutable registry with lookup by key/target/credentialScope.
- **`StorageBackend<K,V>`** — Interface: `get`, `put`, `delete`, `list`, `load`, `flush`, `clear`. Implementations: `InMemoryStorage`, `PersistentStorage` (JSON file), `HybridStorage` (memory + periodic flush), `WalStorage` (write-ahead log + compaction).
- **`AccountAwareStorageBackend`** — Decorator namespacing keys by account ID.
- **`AwsException`** — Domain error with AWS error code + HTTP status.
- **`Resettable`** — Interface for state reset (nuke endpoint).

---

## 3. Complete Service Inventory (68 AWS Services → Azure Mapping)

### 3.1 Stateless Services — Direct Translation

| # | AWS Service (Floci) | Protocol | Azure Equivalent | Translation |
|---|---|---|---|---|
| 1 | SSM (Parameter Store) | JSON 1.1 | **Azure App Configuration** | Direct |
| 2 | SQS | Query + JSON | **Azure Service Bus Queues / Storage Queues** | Direct |
| 3 | SNS | Query + JSON | **Azure Event Grid / Service Bus Topics** | Direct |
| 4 | IAM | Query | **Azure AD / Entra ID (RBAC)** | Adaptation needed |
| 5 | STS | Query | **Azure AD Token Service** | Adaptation needed |
| 6 | KMS | JSON 1.1 | **Azure Key Vault (Keys)** | Direct |
| 7 | Secrets Manager | JSON 1.1 | **Azure Key Vault (Secrets)** | Direct |
| 8 | SES | Query + REST | **Azure Communication Services (Email)** | Direct |
| 9 | Cognito | JSON 1.1 | **Azure AD B2C** | Adaptation needed |
| 10 | Kinesis | JSON 1.1 | **Azure Event Hubs** | Direct |
| 11 | EventBridge | JSON 1.1 | **Azure Event Grid** | Direct |
| 12 | Scheduler | JSON 1.1 | **Azure Logic Apps / Timer Triggers** | Adaptation |
| 13 | AppConfig | REST JSON | **Azure App Configuration (Feature Flags)** | Direct |
| 14 | CloudWatch Logs | JSON 1.1 | **Azure Monitor Logs** | Direct |
| 15 | CloudWatch Metrics | Query + JSON | **Azure Monitor Metrics** | Direct |
| 16 | Step Functions | JSON 1.1 | **Azure Durable Functions / Logic Apps** | Adaptation |
| 17 | CloudFormation | Query | **Azure Resource Manager (ARM/Bicep)** | Major adaptation |
| 18 | ACM | JSON 1.1 | **Azure Key Vault (Certificates)** | Direct |
| 19 | Config | JSON 1.1 | **Azure Policy / Resource Graph** | Adaptation |
| 20 | CloudTrail | JSON 1.1 | **Azure Activity Log / Monitor** | Direct |
| 21 | API Gateway v1 | REST JSON | **Azure API Management** | Direct |
| 22 | API Gateway v2 | REST JSON | **Azure API Management** | Direct |
| 23 | AppSync | REST JSON | **Azure API for GraphQL (APIM)** | Adaptation |
| 24 | ELB v2 | Query | **Azure Load Balancer / App Gateway** | Direct |
| 25 | Auto Scaling | Query | **Azure VM Scale Sets** | Direct |
| 26 | Elastic Beanstalk | Query | **Azure App Service** | Direct |
| 27 | CodeDeploy | JSON 1.1 | **Azure DevOps Pipelines (Release)** | Adaptation |
| 28 | CodePipeline | JSON 1.1 | **Azure DevOps Pipelines** | Adaptation |
| 29 | Backup | JSON REST | **Azure Backup** | Direct |
| 30 | Route53 | REST XML | **Azure DNS** | Direct |
| 31 | Transfer (SFTP) | JSON 1.1 | **Azure Blob Storage SFTP** | Direct |
| 32 | Glue | JSON 1.1 | **Azure Data Factory / Synapse** | Adaptation |
| 33 | Athena | JSON 1.1 | **Azure Synapse SQL** | Adaptation |
| 34 | Pipes | JSON 1.1 | **Azure Functions Bindings** | Adaptation |
| 35 | Firehose | JSON 1.1 | **Azure Stream Analytics** | Adaptation |
| 36 | Textract | JSON 1.1 | **Azure AI Document Intelligence** | Direct |
| 37 | Transcribe | JSON 1.1 | **Azure AI Speech** | Direct |
| 38 | Pricing | JSON 1.1 | **Azure Retail Prices API** | Direct |
| 39 | Cost Explorer | JSON 1.1 | **Azure Cost Management** | Direct |
| 40 | CUR | JSON 1.1 | **Azure Cost Management Exports** | Direct |
| 41 | BCM Data Exports | JSON 1.1 | **Azure Cost Management Exports** | Direct |
| 42 | Resource Groups Tagging | JSON 1.1 | **Azure Resource Graph (tags)** | Direct |
| 43 | Cloud Control | JSON 1.1 | **ARM Generic Provider** | Adaptation |
| 44 | CloudFront | REST XML | **Azure CDN / Front Door** | Direct |
| 45 | Cloud Map | JSON 1.1 | **Azure Service Fabric Naming** | Adaptation |
| 46 | WAFv2 | JSON 1.1 | **Azure WAF (App Gateway/Front Door)** | Direct |
| 47 | Bedrock Runtime | REST JSON | **Azure OpenAI Service** | Direct |
| 48 | Batch | JSON REST | **Azure Batch** | Direct |
| 49 | IoT / IoT Data | REST + MQTT | **Azure IoT Hub** | Direct |
| 50 | Lightsail | JSON 1.1 | *(No direct equivalent — skip)* | Azure-specific |
| 51 | EMR | JSON 1.1 | **Azure HDInsight / Synapse Spark** | Adaptation |

### 3.2 Stateful/Docker-Backed Services

| # | AWS Service (Floci) | Docker Image Used | Azure Equivalent | Translation |
|---|---|---|---|---|
| 52 | S3 | In-process | **Azure Blob Storage** (use Azurite) | Direct |
| 53 | DynamoDB | In-process (H2) | **Azure Cosmos DB (Table API)** | Direct |
| 54 | Lambda | Real Docker containers | **Azure Functions** | Direct (Docker) |
| 55 | RDS (Postgres/MySQL) | postgres/mysql Docker | **Azure SQL / Azure DB for PostgreSQL** | Direct (Docker) |
| 56 | RDS Data API | Proxied to RDS containers | **Azure SQL REST API** | Direct |
| 57 | ElastiCache (Redis) | Redis Docker | **Azure Cache for Redis** | Direct (Docker) |
| 58 | ElastiCache (Memcached) | Memcached Docker | *(No managed Azure equivalent)* | Skip |
| 59 | MemoryDB | Redis Docker | **Azure Cache for Redis (Enterprise)** | Direct |
| 60 | Neptune | TinkerPop/Gremlin Docker | **Azure Cosmos DB (Gremlin API)** | Adaptation |
| 61 | DocumentDB | MongoDB Docker | **Azure Cosmos DB (MongoDB API)** | Direct |
| 62 | OpenSearch | OpenSearch Docker | **Azure AI Search** | Adaptation |
| 63 | MSK (Kafka) | Kafka Docker | **Azure Event Hubs (Kafka protocol)** | Direct |
| 64 | ECS | Docker containers | **Azure Container Instances** | Direct (Docker) |
| 65 | EC2 | Docker containers | **Azure VMs** (emulated) | Adaptation |
| 66 | EKS | K3s/kind Docker | **Azure Kubernetes Service (AKS)** | Adaptation |
| 67 | ECR | In-process registry | **Azure Container Registry** | Direct |
| 68 | CodeBuild | Docker execution | **Azure DevOps Build** | Adaptation |
| 69 | AmazonMQ | RabbitMQ Docker | **Azure Service Bus** | Adaptation |
| 70 | S3 Vectors | In-process | *(Azure-specific: Cosmos DB vector)* | New |

---

## 4. Configuration System

### 4.1 Floci Config Hierarchy
```
EmulatorConfig (Quarkus @ConfigMapping, prefix "floci")
├── port (4566)
├── baseUrl, hostname, defaultRegion, defaultAccountId
├── storage: { mode, persistentPath, hostPersistentPath, pruneVolumesOnDelete, wal, efs, services.* }
├── dns: { extraSuffixes, containerFallbackEnabled, containerFallbackServers }
├── auth: { enabled, iamEnforcement }
├── security: { cors settings }
├── tls: { enabled, certFile, keyFile }
├── initHooks: { enabled, scriptsPath, hooks[] }
├── docker: { socketPath, network, imageRegistryBase, registryCredentials[] }
├── protocols: { strictClaiming }
└── services: { <service>.enabled, <service>.* (per-service config) }
```

### 4.2 floc-zure Config Mapping
| Floci Config | floc-zure Equivalent | Env Var |
|---|---|---|
| `floci.port` | `flocz.port` | `FLOCZ_PORT` |
| `floci.base-url` | `flocz.base-url` | `FLOCZ_BASE_URL` |
| `floci.default-region` | `flocz.default-region` | `FLOCZ_DEFAULT_REGION` (e.g. `eastus`) |
| `floci.default-account-id` | `flocz.default-subscription-id` | `FLOCZ_DEFAULT_SUBSCRIPTION_ID` |
| `floci.storage.mode` | `flocz.storage.mode` | `FLOCZ_STORAGE_MODE` |
| `floci.services.<svc>.enabled` | `flocz.services.<svc>.enabled` | `FLOCZ_SERVICES_<SVC>_ENABLED` |
| `floci.docker.*` | `flocz.docker.*` | `FLOCZ_DOCKER_*` |
| `floci.tls.*` | `flocz.tls.*` | `FLOCZ_TLS_*` |

---

## 5. Build System & CI/CD

| Aspect | Floci | floc-zure (Go) |
|---|---|---|
| Language | Java 25 | Go 1.22+ |
| Framework | Quarkus 3.36 (JAX-RS, Vert.x, CDI) | stdlib `net/http` + Cobra CLI |
| Build tool | Maven (`mvnw`) | `go build` / `Makefile` |
| Container | `eclipse-temurin:25-jre-alpine` | `gcr.io/distroless/static` or `alpine` |
| Native image | GraalVM native-image (~40MB) | Go static binary (~15-30MB) |
| CI | GitHub Actions: `mvn test` | GitHub Actions: `go test ./...` |
| Release | semantic-release (`.releaserc.json`) | goreleaser |
| Compat tests | bats (awscli), Go, Java, Node, Python SDK suites | bats (az cli), Go SDK tests |

---

## 6. Testing Strategy

| Test Type | Floci | floc-zure |
|---|---|---|
| Unit tests | `*ServiceTest.java` (JUnit 5) | `*_test.go` (Go `testing`) |
| Integration tests | `*IntegrationTest.java` (Quarkus test, RestAssured) | `*_integration_test.go` (httptest) |
| SDK compat tests | `compatibility-tests/sdk-test-{awscli,go,java,node,python}/` | `compatibility-tests/sdk-test-{azcli,go}/` |
| IaC compat tests | Terraform, CDK, OpenTofu suites | Terraform (azurerm), Bicep |
| E2E | Docker-based full stack tests | Docker-based full stack tests |

---

## 7. Dependencies & Third-Party Libraries

| Floci Dependency | Purpose | Go Equivalent |
|---|---|---|
| Quarkus (JAX-RS, CDI, Vert.x) | Web framework, DI, reactive | `net/http`, `wire`/manual DI |
| Jackson | JSON/YAML/CBOR serialization | `encoding/json`, `gopkg.in/yaml.v3` |
| docker-java | Docker container management | `github.com/docker/docker/client` |
| BouncyCastle | Crypto (TLS certs, KMS) | `crypto/tls`, `crypto/x509` |
| Apache Kafka client | MSK/Kafka polling | `github.com/IBM/sarama` |
| Vert.x MQTT | IoT MQTT broker | `github.com/mochi-mqtt/server` |
| H2 MVStore | DynamoDB storage engine | `go.etcd.io/bbolt` or badger |
| cron-utils | Cron expression parsing | `github.com/robfig/cron/v3` |
| GraphQL Java | AppSync schema execution | `github.com/graphql-go/graphql` |
| Apache Velocity | AppSync VTL templates | Go template engine |
| Apache MIME4J | SES raw email parsing | `net/mail` |
| Apicurio Registry | Schema registry (Glue) | Custom or skip |

---

## 8. Phased Implementation Plan for floc-zure

### Phase 1: Foundation (Weeks 1-2)
- Go project scaffold (Cobra CLI, config, router, storage interfaces)
- Single-port HTTP server with Azure REST API routing
- Config system (`flocz.yaml` + env vars)
- Storage backends (InMemory + Persistent JSON)
- Health/info/reset endpoints
- Docker integration layer
- `azlocal` CLI wrapper for `az` CLI

### Phase 2: Core Storage Services (Weeks 3-4)
- **Azure Blob Storage** (S3 equivalent — or integrate Azurite)
- **Azure Cosmos DB** (DynamoDB/DocumentDB/Neptune equivalent)
- **Azure Key Vault** (KMS + Secrets Manager + ACM)

### Phase 3: Messaging & Eventing (Weeks 5-6)
- **Azure Service Bus** (SQS + SNS equivalent)
- **Azure Event Grid** (EventBridge equivalent)
- **Azure Event Hubs** (Kinesis equivalent)

### Phase 4: Compute (Weeks 7-8)
- **Azure Functions** (Lambda equivalent — Docker-backed)
- **Azure Container Instances** (ECS equivalent)
- **Azure App Configuration** (SSM + AppConfig)

### Phase 5: Database Services (Weeks 9-10)
- **Azure SQL Database** (RDS equivalent — Docker Postgres/MySQL)
- **Azure Cache for Redis** (ElastiCache equivalent — Docker Redis)
- **Azure DNS** (Route53 equivalent)

### Phase 6: Identity & Management (Weeks 11-12)
- **Azure AD/Entra ID stubs** (IAM/STS equivalent)
- **Azure Monitor** (CloudWatch equivalent)
- **Azure Resource Manager** (CloudFormation equivalent — ARM/Bicep)

### Phase 7: Advanced Services (Weeks 13-14)
- **Azure API Management** (API Gateway)
- **Azure DevOps stubs** (CodePipeline/CodeBuild)
- **Azure AI Services stubs** (Bedrock/Textract/Transcribe)

### Phase 8: Polish & Release (Weeks 15-16)
- E2E test suite (az cli + Go SDK)
- Terraform azurerm provider compatibility
- Documentation, goreleaser, Docker Hub publish
- Performance benchmarks

---

## 9. Priority Services for MVP (Phase 1-3)

Based on developer usage frequency, the MVP should cover:

| Priority | Azure Service | Why |
|---|---|---|
| P0 | Blob Storage | Most used Azure service |
| P0 | Cosmos DB | NoSQL is critical for dev |
| P0 | Service Bus | Messaging is essential |
| P0 | Key Vault | Secrets/keys needed everywhere |
| P0 | Functions | Serverless compute |
| P1 | Event Grid | Event-driven architectures |
| P1 | Azure SQL | Relational DB |
| P1 | App Configuration | Config/feature flags |
| P1 | Cache for Redis | Caching layer |
| P2 | Container Instances | Container workloads |
| P2 | Azure Monitor | Observability |
| P2 | Azure DNS | DNS management |
| P2 | Entra ID stubs | Auth flows |

---

## 10. Key Architectural Decisions for floc-zure

1. **Single binary, single port** — Like Floci, all services on one port (default `10566`). Route by Azure REST API path patterns (`/subscriptions/{subId}/resourceGroups/{rg}/providers/Microsoft.Storage/...`).

2. **Azure REST API wire compatibility** — Emulate the real Azure Resource Manager REST API, so `az` CLI and Azure SDKs work unmodified. Set `AZURE_ENDPOINT_URL` or use `az cloud register`.

3. **Go stdlib HTTP** — No heavy frameworks. Use `net/http` + `gorilla/mux` or `chi` for routing.

4. **Cobra CLI** — `flocz start`, `flocz env`, `flocz stop`, `flocz status`.

5. **Storage interface** — Same 4-mode pattern as Floci (memory/persistent/hybrid/wal).

6. **Docker for stateful services** — Azure Functions, Azure SQL, Redis, Cosmos DB can use real Docker containers like Floci does.

7. **Azurite integration option** — For Blob/Queue/Table storage, optionally delegate to Microsoft's official Azurite emulator running as a sidecar.

8. **Subscription/tenant isolation** — Equivalent to Floci's account isolation, namespace by subscription ID.
