# azfloci — Strategic Improvement Roadmap

> **Based on:** codebase analysis of 14,445 LOC Go, 35 test files, 17 completed phases,
> CI matrix (2 OS × 2 Go versions), plugin architecture, Azure simulator with JSON-file state.

---

## Current State Summary

| Metric | Value | Assessment |
|---|---|---|
| Total Go LOC | 14,445 | Moderate — approaching complexity threshold |
| Test files / Source files | 35 / 42 | Good ratio, but coverage is uneven |
| Highest coverage | `workflow` 98%, `config` 97% | Excellent |
| Lowest coverage | `azure/executor` 11.8%, `cli` 27.5% | **Risk: core integration layer under-tested** |
| State store (`store.go`) | 2,269 lines, single file | **God file — decomposition needed** |
| CI pipeline | lint → test → build → e2e | Solid foundation, missing security/perf gates |
| Storage backend | Single JSON file + mutex | Works for dev; not viable at scale |
| Dependency count | 13 (all indirect except cobra/viper/yaml) | Lean — good |

---

## 1. Architecture & Scalability

### 1.1 Decompose the State Store God File
| | |
|---|---|
| **WHY** | `simulator/internal/state/store.go` is 2,269 lines handling subscriptions, resource groups, storage accounts, key vaults, networking, VMs, service bus, and functions — all behind a single `sync.RWMutex`. Any new Azure service adds to the monolith. |
| **WHAT** | Split into domain-specific stores (`store_compute.go`, `store_network.go`, `store_keyvault.go`, etc.) behind a `StateManager` facade. Each domain store owns its own lock scope. Test files already follow this pattern (`store_keyvault_test.go`, `store_vm_test.go`) — the implementation should match. |
| **WHEN** | **Near-term (next phase)** — complexity is already at the threshold. Every new Azure service makes this harder. |

### 1.2 Storage Backend Abstraction
| | |
|---|---|
| **WHY** | The `FLOCI_ANALYSIS.md` blueprint calls for 4 storage modes (in-memory, persistent, hybrid, WAL) matching the Floci reference architecture, but only JSON-file persistence exists today. This limits throughput and crash resilience. |
| **WHAT** | Implement the `StorageBackend` interface from the blueprint: `InMemoryStore` (fast tests, ephemeral), `HybridStore` (memory + periodic flush), `WALStore` (write-ahead log for crash recovery). Let users select via `AZFLOCI_STORAGE_MODE`. |
| **WHEN** | **Mid-term** — after god-file decomposition makes it feasible. |

### 1.3 Event-Driven Internals for Cross-Service Triggers
| | |
|---|---|
| **WHY** | Azure services interact (e.g., Service Bus triggers Functions, storage events trigger Event Grid). Currently each handler is isolated — no event bus. |
| **WHAT** | Add an internal `EventBus` (in-process pub/sub with Go channels). Service handlers publish events (`BlobCreated`, `MessageEnqueued`); consumers (Functions, Event Grid) subscribe. This mirrors Azure's real behavior and enables realistic integration testing. |
| **WHEN** | **Mid-term** — after Functions and Service Bus stabilize. |

### 1.4 API Versioning Strategy
| | |
|---|---|
| **WHY** | Azure REST API is version-stamped (`?api-version=2023-01-01`). The simulator currently ignores versions, which will cause drift as users target specific API behaviors. |
| **WHAT** | Route handlers should accept and validate `api-version` query params. Start with "latest wins" semantics but log warnings for unknown versions. Add a `--strict-api-version` flag that rejects unrecognized versions. |
| **WHEN** | **Mid-term** — before external users adopt the simulator. |

---

## 2. Developer Experience

### 2.1 Close the Coverage Gaps in Core Layers
| | |
|---|---|
| **WHY** | `internal/azure` (11.8%) is the CLI executor — the literal bridge between user commands and Azure. `internal/cli` (27.5%) is the command layer. Bugs here are user-facing and hard to catch. |
| **WHAT** | Add interface-driven mocking for `CLIExecutor` (the `CONTRIBUTING.md` already mentions this pattern but it's under-used). Target 70%+ on both packages. Add table-driven tests for every `az` subcommand path. |
| **WHEN** | **Near-term (next phase)** — these are the highest-risk low-coverage modules. |

### 2.2 CI Pipeline Enhancements
| | |
|---|---|
| **WHY** | Current CI runs lint → test → build → e2e. Missing: coverage enforcement, security scanning, Docker image validation, release automation. Coverage regressions can slip in silently. |
| **WHAT** | (a) Add `go tool cover` threshold check (fail CI if total coverage drops below 60%). (b) Add `govulncheck` step for dependency vulnerability scanning. (c) Add Docker build+smoke in CI. (d) Wire `.goreleaser.yml` (already exists) to a release workflow on tag push. |
| **WHEN** | **Near-term** — low effort, high signal. |

### 2.3 Local Dev Environment: `make dev` One-Command Setup
| | |
|---|---|
| **WHY** | New contributors must read `CONTRIBUTING.md`, run `go mod tidy`, `make build`, understand the simulator relationship. Friction slows adoption. |
| **WHAT** | Add `make dev` target that: builds both binaries, starts simulator in background, sets env vars, runs smoke test, prints "ready" URL. Add a `devcontainer.json` for Codespaces/VS Code with Go tooling pre-installed. |
| **WHEN** | **Near-term** — contributor funnel optimization. |

### 2.4 Contract Testing Between CLI and Simulator
| | |
|---|---|
| **WHY** | CLI and simulator are loosely coupled through `az` CLI output JSON format. If the simulator changes its output schema, CLI tests still pass but real usage breaks. |
| **WHAT** | Define JSON Schema contracts for each command's output (e.g., `az group show` response shape). Both CLI parsers and simulator handlers validate against the same schema. CI runs contract validation. |
| **WHEN** | **Mid-term** — after the output surface stabilizes. |

---

## 3. Observability & Reliability

### 3.1 Structured Logging
| | |
|---|---|
| **WHY** | The codebase uses `fmt.Errorf` and ad-hoc error returns. No structured logging means debugging simulator issues requires reading code, not logs. |
| **WHAT** | Adopt `log/slog` (stdlib, Go 1.21+). Add request-scoped fields: `command`, `resource_type`, `subscription_id`, `duration_ms`. Emit JSON logs in simulator mode, human-readable in CLI mode. |
| **WHEN** | **Near-term** — minimal effort with Go's built-in `slog`. |

### 3.2 Simulator Request Tracing
| | |
|---|---|
| **WHY** | When the simulator handles complex workflows (ARM deployments, drift detection across many resources), there's no way to trace which operations occurred or their order. |
| **WHAT** | Add a `/debug/trace` endpoint on the simulator that returns the last N operations with timestamps, resource paths, and outcomes. Optionally support OpenTelemetry span export for integration with external tracing (Jaeger, Zipkin). |
| **WHEN** | **Mid-term** — after structured logging is in place. |

### 3.3 Health Check and Readiness Probes
| | |
|---|---|
| **WHY** | Docker Compose exposes port 8080 ("reserved for future HTTP API") but no health endpoint exists. Container orchestrators can't verify readiness. |
| **WHAT** | Add `/healthz` (liveness) and `/readyz` (state loaded, ready to serve) endpoints. Wire into `docker-compose.yml` healthcheck directive. |
| **WHEN** | **Near-term** — 20 lines of code, enables reliable container orchestration. |

### 3.4 Define SLOs for Simulator Fidelity
| | |
|---|---|
| **WHY** | Users need to know "how Azure-like is this simulator?" No fidelity metrics exist. |
| **WHAT** | Define SLIs: (a) % of `az` CLI commands that produce identical output to real Azure, (b) % of ARM template resources that deploy successfully, (c) response time per command. Track in a compatibility matrix (`docs/compatibility.md`). Run periodic fidelity tests against a real Azure subscription. |
| **WHEN** | **Long-term** — requires real Azure baseline. |

---

## 4. Security

### 4.1 Dependency Vulnerability Scanning
| | |
|---|---|
| **WHY** | 13 dependencies (direct+indirect). No automated scanning. A vulnerability in `cobra`, `viper`, or `afero` would go unnoticed. |
| **WHAT** | Add `govulncheck ./...` to CI. Add Dependabot or Renovate for automated dependency PRs. Pin Go version in `go.mod` and Dockerfile. |
| **WHEN** | **Near-term** — 5 minutes to add, catches real risks. |

### 4.2 Secrets Management for Test Fixtures
| | |
|---|---|
| **WHY** | The simulator handles Key Vault secrets. Test fixtures may include dummy credentials that could be confused with real ones, or future integration tests may need real Azure credentials. |
| **WHAT** | (a) Add `.gitguardian.yml` or `gitleaks` to CI to prevent accidental secret commits. (b) For future integration tests, use GitHub Actions secrets + OIDC federation (no long-lived credentials). |
| **WHEN** | **Near-term** — shift-left, prevent incidents before they happen. |

### 4.3 Input Validation and Fuzzing
| | |
|---|---|
| **WHY** | The simulator accepts arbitrary JSON in resource properties (`map[string]interface{}`). Malformed input could cause panics, infinite loops, or state corruption. |
| **WHAT** | Add Go native fuzzing (`go test -fuzz`) for the ARM template parser and JSON state deserializer. Add input size limits to the simulator HTTP handler. |
| **WHEN** | **Mid-term** — after core stability improves. |

### 4.4 RBAC Simulation for Zero-Trust Testing
| | |
|---|---|
| **WHY** | The blueprint maps IAM/STS → Azure AD/Entra ID. Currently the simulator accepts all operations — no auth/authz simulation. Users can't test permission-denied scenarios. |
| **WHAT** | Add opt-in RBAC simulation: define roles/permissions in config, reject operations that lack required role. Start with Reader/Contributor/Owner built-in roles. |
| **WHEN** | **Long-term** — complex, but high value for enterprise users. |

---

## 5. Process & Collaboration

### 5.1 Architecture Decision Records (ADRs)
| | |
|---|---|
| **WHY** | 17 phases completed with no recorded architectural decisions. Why JSON-file state instead of SQLite? Why Cobra over alternatives? Why Go over the Floci Java approach? Future contributors will re-litigate these decisions. |
| **WHAT** | Create `docs/adr/` directory. Retroactively document the top 5 decisions: (1) Go as implementation language, (2) single-binary architecture, (3) az CLI wire compatibility, (4) JSON-file state persistence, (5) plugin provider pattern. Use the MADR template. |
| **WHEN** | **Near-term** — knowledge preservation, low effort. |

### 5.2 Tech Debt Tracking with Labels
| | |
|---|---|
| **WHY** | The god-file, low coverage areas, and missing storage backends are implicit tech debt. No systematic tracking means prioritization is ad-hoc. |
| **WHAT** | Create GitHub issue labels: `tech-debt`, `coverage-gap`, `refactor`. File issues for known debt items. Reference them in sprint planning. Add a `TECH_DEBT.md` with current inventory and severity ratings. |
| **WHEN** | **Near-term** — process, not code. |

### 5.3 Feature Flags for Experimental Services
| | |
|---|---|
| **WHY** | New Azure service simulations (e.g., Cosmos DB, Azure SQL) may be incomplete. Shipping them without a way to disable them risks breaking stable workflows. |
| **WHAT** | Add per-service enable/disable flags in config (`services.cosmosdb.enabled: false`). The Floci blueprint already specifies `ServiceDescriptor.enabled` — implement it. Default new services to disabled until they pass a fidelity threshold. |
| **WHEN** | **Mid-term** — before adding the next wave of Azure services. |

### 5.4 Changelog Automation
| | |
|---|---|
| **WHY** | `CHANGELOG.md` exists but requires manual updates. Conventional commits are specified in `CONTRIBUTING.md` but not enforced or leveraged. |
| **WHAT** | Add `commitlint` to CI (enforce conventional commits). Use `goreleaser` changelog generation from commit messages. Add a `release` GitHub Actions workflow triggered by version tags. |
| **WHEN** | **Near-term** — pairs with the `.goreleaser.yml` already in the repo. |

---

## 6. Performance

### 6.1 Benchmark Suite for Simulator Operations
| | |
|---|---|
| **WHY** | No benchmarks exist. The single-mutex state store may become a bottleneck as resource counts grow. No baseline means no way to detect regressions. |
| **WHAT** | Add Go benchmarks (`BenchmarkCreateResource`, `BenchmarkListResources_1000`, `BenchmarkPersistState`) using `testing.B`. Run in CI with `benchstat` to detect regressions (>10% slowdown fails the build). |
| **WHEN** | **Near-term** — establishes baselines before architectural changes. |

### 6.2 State Persistence Optimization
| | |
|---|---|
| **WHY** | Every mutation calls `persist()` which marshals the *entire* state to JSON and writes the full file. With 1,000+ resources, this becomes O(n) per write. |
| **WHAT** | (a) **Near-term**: batch writes with a debounce (persist at most once per 100ms). (b) **Mid-term**: implement the WAL backend — append-only writes, periodic compaction. (c) **Long-term**: consider SQLite via `modernc.org/sqlite` (pure Go, no CGO) for indexed queries. |
| **WHEN** | **Phased: near → mid → long-term** as described above. |

### 6.3 Concurrent Handler Safety Audit
| | |
|---|---|
| **WHY** | The simulator uses a single `sync.RWMutex` for all state. This serializes all write operations across all resource types. Under concurrent CLI usage (e.g., parallel Terraform apply), this is a bottleneck. |
| **WHAT** | After decomposing the state store (§1.1), each domain store gets its own lock. Add `-race` detector to all CI test runs (already present — good). Add a concurrent stress test: 100 goroutines creating/listing/deleting resources simultaneously. |
| **WHEN** | **Mid-term** — depends on §1.1 completion. |

### 6.4 CLI Command Latency Profiling
| | |
|---|---|
| **WHY** | CLI startup time matters for developer experience. Cobra + Viper initialization, config loading, and executor setup all contribute. No profiling data exists. |
| **WHAT** | Add `--profile` flag that emits a `pprof` CPU profile. Measure baseline startup time in CI. Target: <100ms for `azfloci version`, <500ms for any simulator command. |
| **WHEN** | **Long-term** — optimization, not a current bottleneck. |

---

## Prioritized Implementation Sequence

### Phase Next (Near-Term) — Do First
| # | Item | Effort | Impact |
|---|---|---|---|
| 1 | State store decomposition (§1.1) | Medium | Unblocks everything else |
| 2 | Coverage gaps: `azure` + `cli` (§2.1) | Medium | Risk reduction |
| 3 | CI: coverage gate + `govulncheck` + release workflow (§2.2) | Low | Quality gate |
| 4 | Structured logging with `slog` (§3.1) | Low | Debuggability |
| 5 | Health endpoints (§3.3) | Low | Container readiness |
| 6 | ADRs for top 5 decisions (§5.1) | Low | Knowledge preservation |
| 7 | Benchmark suite (§6.1) | Low | Baseline before refactors |
| 8 | `gitleaks` + `govulncheck` in CI (§4.1, §4.2) | Low | Shift-left security |
| 9 | Changelog automation (§5.4) | Low | Release process |

### Phase Next+1 (Mid-Term) — Build On Foundation
| # | Item | Effort | Impact |
|---|---|---|---|
| 10 | Storage backend abstraction (§1.2) | High | Scalability |
| 11 | Event bus for cross-service triggers (§1.3) | High | Realism |
| 12 | API version validation (§1.4) | Medium | Compatibility |
| 13 | Contract testing (§2.4) | Medium | Integration safety |
| 14 | Request tracing (§3.2) | Medium | Observability |
| 15 | Input fuzzing (§4.3) | Medium | Robustness |
| 16 | Feature flags for services (§5.3) | Medium | Safe rollout |
| 17 | WAL storage + write batching (§6.2) | Medium | Performance |
| 18 | Concurrent stress tests (§6.3) | Medium | Reliability |

### Phase Next+2 (Long-Term) — Strategic Investments
| # | Item | Effort | Impact |
|---|---|---|---|
| 19 | Fidelity SLOs + Azure baseline tests (§3.4) | High | Trust & adoption |
| 20 | RBAC simulation (§4.4) | High | Enterprise readiness |
| 21 | SQLite state backend (§6.2c) | High | Query performance |
| 22 | CLI latency profiling (§6.4) | Low | DX polish |

---

*Generated from analysis of `floc-zure` at 17 completed phases, 14,445 LOC.*
