# Future Phase Improvements

> **Project:** azfloci (floc-zure) — Azure Cloud Simulator & CLI  
> **Template Version:** 1.0  
> **Last Updated:** <!-- YYYY-MM-DD -->  
> **Owner:** <!-- Team / Individual -->  
> **Status:** Draft | In Review | Approved  
> **Review Cadence:** Quarterly (see §8)

---

## How to Use This Document

This template captures improvement suggestions across **technical**, **process**, and **people** dimensions, prioritizes them using an impact-vs-effort matrix, and sequences them into a phased roadmap. It is a living document — update it each review cycle.

**Conventions:**
- `<!-- PLACEHOLDER -->` — replace with project-specific content.
- `⚑ GUIDANCE` blocks — instruction callouts; delete once filled.
- Example entries are provided in *italics* — replace or extend with real items.
- Priority: P1 (critical) → P4 (nice-to-have).

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State Assessment](#2-current-state-assessment)
3. [Improvement Categories](#3-improvement-categories)
4. [Prioritization Matrix](#4-prioritization-matrix)
5. [Phased Roadmap](#5-phased-roadmap)
6. [Success Metrics & KPIs](#6-success-metrics--kpis)
7. [Risks & Dependencies](#7-risks--dependencies)
8. [Review Cadence](#8-review-cadence)
9. [Appendix: Decision Log](#appendix-a-decision-log)

---

## 1. Executive Summary

> ⚑ GUIDANCE: 3–5 sentences. State the project's current maturity, the strategic
> theme for upcoming improvements, and the expected outcome. Write this last,
> after filling all other sections.

<!-- EXECUTIVE SUMMARY START -->

*azfloci has completed 17 phases delivering a fully functional Azure CLI wrapper with
a cloud simulator supporting 8 Azure service families, drift detection, ARM template
deployment, plugin architecture, and cost estimation across 14,445 lines of Go.
The codebase is functionally complete for its MVP scope but carries architectural
debt (2,269-line state store monolith, uneven test coverage in core integration layers)
and lacks production-readiness features (structured logging, storage backend
abstraction, API version validation). This document identifies 22 improvement
opportunities organized into three phases spanning the next 3–4 quarters, targeting
maintainability, scalability, and developer experience.*

<!-- EXECUTIVE SUMMARY END -->

---

## 2. Current State Assessment

### 2.1 Quantitative Snapshot

> ⚑ GUIDANCE: Fill in key metrics from your project. Add or remove rows as needed.
> Use CI reports, `go tool cover`, `cloc`, and dependency audit tools as data sources.

| Metric | Current Value | Target | Gap | Source |
|--------|--------------|--------|-----|--------|
| Total LOC | *14,445* | *—* | *—* | *`cloc`* |
| Test-to-source file ratio | *35 / 42* | *≥ 0.8* | *0.83 ✓* | *`find . -name '*_test.go'`* |
| Highest module coverage | *workflow 98%, config 97%* | *≥ 90%* | *Met ✓* | *`go tool cover`* |
| Lowest module coverage | *azure/executor 11.8%, cli 27.5%* | *≥ 70%* | *⚠ 58pp / 42pp gap* | *`go tool cover`* |
| Largest single file | *store.go — 2,269 lines* | *< 500 lines* | *⚠ 4.5× over threshold* | *`wc -l`* |
| Direct dependencies | *13* | *< 20* | *Met ✓* | *`go mod graph`* |
| CI gate stages | *4 (lint → test → build → e2e)* | *6+* | *Missing security, perf* | *`.github/workflows/`* |
| <!-- Metric --> | <!-- Value --> | <!-- Target --> | <!-- Gap --> | <!-- Source --> |

### 2.2 Pain Points

> ⚑ GUIDANCE: List concrete pain points experienced by the team. Each entry needs
> an **impact** (who is affected, how badly) and **evidence** (bug reports, slow
> CI times, developer complaints, incident postmortems).

| # | Pain Point | Impact | Evidence | Category |
|---|-----------|--------|----------|----------|
| PP-1 | *State store monolith (`store.go` = 2,269 LOC)* | *Every new Azure service increases file complexity; merge conflicts frequent* | *Test files already split by domain but impl is not* | Technical |
| PP-2 | *Core integration layer under-tested (executor 11.8%)* | *User-facing bugs escape to production; CLI regressions* | *Coverage reports; no mock-based executor tests* | Technical |
| PP-3 | *No structured logging* | *Debugging simulator issues requires reading source code* | *Only `fmt.Errorf` used; no log levels or request tracing* | Technical |
| PP-4 | *Manual release process* | *GoReleaser config exists but no automated tag-triggered workflow* | *`.goreleaser.yml` present, no release workflow* | Process |
| PP-5 | *New contributor onboarding friction* | *Must read multiple docs, run several commands to get started* | *No `make dev` target or devcontainer* | People |
| PP-6 | <!-- Pain Point --> | <!-- Impact --> | <!-- Evidence --> | <!-- Category --> |

### 2.3 Bottlenecks

> ⚑ GUIDANCE: Bottlenecks are constraints that limit throughput, scalability, or
> velocity. Distinguish from pain points: a bottleneck is the *binding constraint*
> that, if resolved, unlocks the most capacity.

| # | Bottleneck | Constraint Type | Current Throughput | Desired Throughput |
|---|-----------|----------------|-------------------|-------------------|
| BN-1 | *Single JSON file + mutex for all state* | *I/O & concurrency* | *Adequate for dev; blocks multi-user & scale testing* | *Concurrent reads, crash-resilient writes* |
| BN-2 | *No event bus between service handlers* | *Architecture* | *Services are isolated; no cross-service triggers* | *Service Bus triggers Functions; storage events trigger Event Grid* |
| BN-3 | *API version ignored on all routes* | *Compatibility* | *Any SDK version works but behavior may not match real Azure* | *Version-aware routing with strict mode option* |
| BN-4 | <!-- Bottleneck --> | <!-- Type --> | <!-- Current --> | <!-- Desired --> |

### 2.4 Strengths to Preserve

> ⚑ GUIDANCE: Not everything needs fixing. Document what's working well so
> improvements don't accidentally regress these areas.

| # | Strength | Evidence |
|---|---------|----------|
| S-1 | *Lean dependency tree (13 deps)* | *`go mod graph` — only cobra/viper/yaml are direct* |
| S-2 | *Clean interface-driven architecture* | *`CLIExecutor`, `config.Manager`, `resource.Manager`, `workflow.Engine` — all mockable* |
| S-3 | *Excellent coverage in core domain logic* | *workflow 98%, config 97%* |
| S-4 | *Comprehensive CI matrix (2 OS × 2 Go versions)* | *`.github/workflows/ci.yml`* |
| S-5 | *Plugin architecture already in place* | *`ResourceProvider` interface with thread-safe registry* |
| S-6 | <!-- Strength --> | <!-- Evidence --> |


---

## 3. Improvement Categories

> ⚑ GUIDANCE: Organize improvements into three pillars. Each item should reference
> a pain point (PP-#) or bottleneck (BN-#) from §2 where applicable. Use a unique
> ID (T-#, PR-#, PE-#) for cross-referencing in the prioritization matrix and roadmap.

### 3.1 Technical Improvements

Improvements to code, architecture, infrastructure, and tooling.

| ID | Improvement | Description | Addresses | Effort | Impact |
|----|------------|-------------|-----------|--------|--------|
| T-1 | *Decompose state store god file* | *Split `store.go` (2,269 LOC) into domain-specific stores (`store_compute.go`, `store_network.go`, etc.) behind a `StateManager` facade. Each domain store owns its own lock scope.* | *PP-1, BN-1* | Medium | High |
| T-2 | *Close coverage gaps in core layers* | *Add interface-driven mocking for `CLIExecutor`. Target 70%+ on `internal/azure` (11.8% → 70%) and `internal/cli` (27.5% → 70%). Table-driven tests for every `az` subcommand path.* | *PP-2* | Medium | High |
| T-3 | *Adopt structured logging (`log/slog`)* | *Replace `fmt.Errorf` with Go 1.21+ `slog`. Request-scoped fields: command, resource_type, subscription_id, duration_ms. JSON logs in simulator mode, human-readable in CLI mode.* | *PP-3* | Low | High |
| T-4 | *CI pipeline hardening* | *(a) Coverage threshold gate (fail if < 60%). (b) `govulncheck` for dependency vulnerabilities. (c) Docker build + smoke test. (d) Tag-triggered release workflow via GoReleaser.* | *PP-4* | Low | High |
| T-5 | *Storage backend abstraction* | *Implement `StorageBackend` interface: `InMemoryStore`, `HybridStore` (memory + periodic flush), `WALStore` (write-ahead log). Select via `AZFLOCI_STORAGE_MODE` env var.* | *BN-1* | High | High |
| T-6 | *Internal event bus* | *In-process pub/sub with Go channels. Service handlers publish events (`BlobCreated`, `MessageEnqueued`); consumers (Functions, Event Grid) subscribe. Mirrors real Azure behavior.* | *BN-2* | High | High |
| T-7 | *API version validation* | *Route handlers accept/validate `api-version` query param. "Latest wins" default + `--strict-api-version` flag that rejects unknown versions. Log warnings for unrecognized versions.* | *BN-3* | Medium | Medium |
| T-8 | *Contract testing (CLI ↔ Simulator)* | *JSON Schema contracts for each command's output. Both CLI parsers and simulator handlers validate against the same schema. CI runs contract validation.* | *—* | Medium | Medium |
| T-9 | *Request tracing in simulator* | *OpenTelemetry-compatible trace IDs on every simulator request. Correlate multi-resource operations (ARM deployments, drift detection).* | *PP-3* | Medium | Medium |
| T-10 | *Input fuzzing for parsers* | *Go native fuzzing (`testing.F`) for ARM template parser, YAML workflow parser, and CLI argument parsing. Run in CI nightly.* | *—* | Medium | Low |
| T-11 | <!-- Improvement --> | <!-- Description --> | <!-- Addresses --> | <!-- Effort --> | <!-- Impact --> |

### 3.2 Process Improvements

Improvements to workflows, CI/CD, release management, and team practices.

| ID | Improvement | Description | Addresses | Effort | Impact |
|----|------------|-------------|-----------|--------|--------|
| PR-1 | *Automated release workflow* | *Wire `.goreleaser.yml` to a GitHub Actions workflow triggered by `v*` tags. Generate changelogs, publish binaries, create GitHub Release.* | *PP-4* | Low | High |
| PR-2 | *Feature flag system for services* | *Per-service `enabled` toggle in config (e.g., `AZFLOCI_ENABLE_SERVICEBUS=true`). Allows gradual rollout and disabling unstable services.* | *—* | Medium | Medium |
| PR-3 | *Changelog automation* | *Use `git-cliff` or `conventional-changelog` to generate `CHANGELOG.md` entries from conventional commit messages. Enforce commit message format in CI.* | *—* | Low | Low |
| PR-4 | *ADR (Architecture Decision Records)* | *Create `docs/adr/` directory. Template for recording architectural decisions with context, options considered, decision, and consequences.* | *—* | Low | Medium |
| PR-5 | *Dependency update automation* | *Enable Dependabot or Renovate for `go.mod` updates. Weekly PR cadence with auto-merge for patch versions.* | *—* | Low | Medium |
| PR-6 | <!-- Improvement --> | <!-- Description --> | <!-- Addresses --> | <!-- Effort --> | <!-- Impact --> |

### 3.3 People & Culture Improvements

Improvements to onboarding, knowledge sharing, and developer experience.

| ID | Improvement | Description | Addresses | Effort | Impact |
|----|------------|-------------|-----------|--------|--------|
| PE-1 | *`make dev` one-command setup* | *Single Makefile target: builds both binaries, starts simulator in background, sets env vars, runs smoke test, prints "ready" URL. Reduces onboarding from 5 steps to 1.* | *PP-5* | Low | High |
| PE-2 | *DevContainer / Codespaces config* | *`devcontainer.json` with Go tooling, golangci-lint, and simulator pre-installed. Zero-friction cloud development environment.* | *PP-5* | Low | Medium |
| PE-3 | *Architecture walkthrough recording* | *30-minute video or written walkthrough of the codebase architecture for new contributors. Cover: package layout, data flow, extension points.* | *PP-5* | Low | Medium |
| PE-4 | *Contributor office hours* | *Monthly 1-hour open session for contributors to ask questions, pair program, or discuss design decisions.* | *—* | Low | Low |
| PE-5 | <!-- Improvement --> | <!-- Description --> | <!-- Addresses --> | <!-- Effort --> | <!-- Impact --> |

---

## 4. Prioritization Matrix

### 4.1 Impact vs. Effort Grid

> ⚑ GUIDANCE: Plot each improvement ID on the 2×2 matrix below. Items in the
> **top-left (Quick Wins)** should be tackled first. Items in the **bottom-right
> (Money Pit)** should be deferred or dropped.

```
                        ┌─────────────────────────────────────────┐
                        │              IMPACT                     │
                        │     High                    Low         │
                   ┌────┼─────────────────────┬───────────────────┤
                   │    │  🏆 QUICK WINS       │  ⚡ FILL-INS      │
                   │Low │  T-3  (slog)         │  PR-3 (changelog) │
            EFFORT │    │  T-4  (CI harden)    │  PE-4 (office hrs)│
                   │    │  PR-1 (auto release) │                   │
                   │    │  PE-1 (make dev)     │                   │
                   │    │  PE-2 (devcontainer) │                   │
                   │    │  PR-5 (dependabot)   │                   │
                   ├────┼─────────────────────┼───────────────────┤
                   │    │  🎯 BIG BETS         │  💀 MONEY PIT     │
                   │High│  T-1  (decompose)    │  T-10 (fuzzing)   │
                   │    │  T-2  (coverage)     │                   │
                   │    │  T-5  (storage)      │                   │
                   │    │  T-6  (event bus)    │                   │
                   │    │  T-7  (API version)  │                   │
                   │    │  T-8  (contracts)    │                   │
                   │    │  T-9  (tracing)      │                   │
                   └────┼─────────────────────┴───────────────────┤
                        └─────────────────────────────────────────┘
```

### 4.2 Prioritized Ranked List

> ⚑ GUIDANCE: Final prioritization combining impact, effort, risk, and
> dependencies. P1 = must do next; P4 = backlog.

| Rank | ID | Title | Priority | Effort | Impact | Dependencies | Target Phase |
|------|-----|-------|----------|--------|--------|-------------|-------------|
| 1 | T-3 | Structured logging (`slog`) | P1 | Low | High | None | N+1 |
| 2 | T-4 | CI pipeline hardening | P1 | Low | High | None | N+1 |
| 3 | PE-1 | `make dev` one-command setup | P1 | Low | High | None | N+1 |
| 4 | PR-1 | Automated release workflow | P1 | Low | High | T-4 | N+1 |
| 5 | T-1 | Decompose state store | P1 | Medium | High | None | N+1 |
| 6 | T-2 | Close coverage gaps | P1 | Medium | High | None | N+1 |
| 7 | PR-4 | Architecture Decision Records | P2 | Low | Medium | None | N+1 |
| 8 | PE-2 | DevContainer / Codespaces | P2 | Low | Medium | None | N+1 |
| 9 | PR-5 | Dependency update automation | P2 | Low | Medium | None | N+1 |
| 10 | T-5 | Storage backend abstraction | P2 | High | High | T-1 | N+2 |
| 11 | T-6 | Internal event bus | P2 | High | High | T-1 | N+2 |
| 12 | T-7 | API version validation | P2 | Medium | Medium | None | N+2 |
| 13 | T-8 | Contract testing | P2 | Medium | Medium | T-2 | N+2 |
| 14 | T-9 | Request tracing | P3 | Medium | Medium | T-3 | N+2 |
| 15 | PR-2 | Feature flags for services | P3 | Medium | Medium | None | N+2 |
| 16 | PE-3 | Architecture walkthrough | P3 | Low | Medium | None | N+2 |
| 17 | T-10 | Input fuzzing | P4 | Medium | Low | None | N+3 |
| 18 | PR-3 | Changelog automation | P4 | Low | Low | PR-1 | N+3 |
| 19 | PE-4 | Contributor office hours | P4 | Low | Low | None | N+3 |


---

## 5. Phased Roadmap

> ⚑ GUIDANCE: Each phase should be completable in ~1 quarter. Phase N+1 focuses
> on quick wins and foundation work. N+2 builds on that foundation. N+3 is
> strategic investment. Adjust timelines to your team's capacity.

### Phase N+1 — Foundation & Quick Wins (Q<!-- next -->)

**Theme:** *Reduce debt, harden CI, improve DX — make the codebase ready for scale.*

**Duration:** 6–8 weeks | **Capacity:** <!-- X --> person-weeks

| # | Item | ID | Owner | Est. Days | Done Criteria | Status |
|---|------|----|-------|-----------|---------------|--------|
| 1 | Adopt `log/slog` across all packages | T-3 | <!-- owner --> | 3 | All `fmt.Errorf` replaced; JSON log output in simulator; human-readable in CLI | ☐ |
| 2 | CI: coverage gate + `govulncheck` + Docker smoke | T-4 | <!-- owner --> | 2 | CI fails on coverage < 60%; zero known vulns; Docker build passes | ☐ |
| 3 | `make dev` one-command setup | PE-1 | <!-- owner --> | 2 | `make dev` builds, starts simulator, runs smoke, prints URL | ☐ |
| 4 | Tag-triggered release workflow | PR-1 | <!-- owner --> | 1 | `git tag v0.6.0 && git push --tags` → GitHub Release with binaries | ☐ |
| 5 | Decompose `store.go` into domain stores | T-1 | <!-- owner --> | 5 | `store.go` < 300 LOC; 7+ domain files; all existing tests pass | ☐ |
| 6 | Executor + CLI coverage to 70%+ | T-2 | <!-- owner --> | 5 | `go tool cover` shows ≥70% on `internal/azure` and `internal/cli` | ☐ |
| 7 | ADR directory + first 3 decisions | PR-4 | <!-- owner --> | 1 | `docs/adr/001-*.md` through `003-*.md` exist | ☐ |
| 8 | DevContainer config | PE-2 | <!-- owner --> | 1 | `.devcontainer/devcontainer.json` works in VS Code / Codespaces | ☐ |
| 9 | Dependabot / Renovate config | PR-5 | <!-- owner --> | 0.5 | `.github/dependabot.yml` or `renovate.json` merges weekly patches | ☐ |
|   | **Phase total** | | | **~20.5** | | |

**Exit Criteria for Phase N+1:**
- [ ] All 9 items marked ☑
- [ ] CI pipeline green with new gates
- [ ] `store.go` decomposed; no regression in test suite
- [ ] Coverage ≥ 70% on previously under-tested modules
- [ ] New contributor can run `make dev` and have a working environment in < 2 minutes

---

### Phase N+2 — Scale & Integration (Q<!-- next+1 -->)

**Theme:** *Storage abstraction, cross-service events, and contract safety — make the simulator production-grade.*

**Duration:** 8–10 weeks | **Capacity:** <!-- X --> person-weeks

| # | Item | ID | Owner | Est. Days | Done Criteria | Status |
|---|------|----|-------|-----------|---------------|--------|
| 1 | Storage backend abstraction | T-5 | <!-- owner --> | 8 | `InMemoryStore`, `HybridStore`, `WALStore` pass integration tests; env var toggle works | ☐ |
| 2 | Internal event bus | T-6 | <!-- owner --> | 8 | `EventBus` with ≥3 event types; Service Bus → Functions trigger works in e2e test | ☐ |
| 3 | API version validation | T-7 | <!-- owner --> | 4 | All routes accept `api-version`; `--strict-api-version` flag works; tests cover both modes | ☐ |
| 4 | Contract testing framework | T-8 | <!-- owner --> | 5 | JSON Schema defined for ≥10 commands; CI validates contracts | ☐ |
| 5 | Request tracing | T-9 | <!-- owner --> | 4 | Trace IDs in all simulator responses; correlation across ARM deployment steps | ☐ |
| 6 | Feature flags for services | PR-2 | <!-- owner --> | 3 | Per-service `enabled` config; disabled services return 501 with clear message | ☐ |
| 7 | Architecture walkthrough doc | PE-3 | <!-- owner --> | 2 | Written guide or video published; linked from `CONTRIBUTING.md` | ☐ |
|   | **Phase total** | | | **~34** | | |

**Exit Criteria for Phase N+2:**
- [ ] Storage backend is selectable and all 3 modes pass benchmarks
- [ ] ≥ 1 cross-service event flow works end-to-end
- [ ] Contract tests prevent schema drift between CLI and simulator
- [ ] New Azure API version can be added without modifying existing handlers

---

### Phase N+3 — Strategic & Polish (Q<!-- next+2 -->)

**Theme:** *Enterprise readiness, robustness, and community growth.*

**Duration:** 8–12 weeks | **Capacity:** <!-- X --> person-weeks

| # | Item | ID | Owner | Est. Days | Done Criteria | Status |
|---|------|----|-------|-----------|---------------|--------|
| 1 | Input fuzzing for all parsers | T-10 | <!-- owner --> | 5 | Fuzz targets for ARM, YAML, CLI arg parsers; nightly CI run; zero crashes after 10M iterations | ☐ |
| 2 | Changelog automation | PR-3 | <!-- owner --> | 1 | `CHANGELOG.md` auto-generated on release; commit format enforced | ☐ |
| 3 | Contributor office hours | PE-4 | <!-- owner --> | — | Monthly cadence established; first 3 sessions held | ☐ |
| 4 | *RBAC simulation* | *—* | <!-- owner --> | 10 | *Role assignments enforced on simulator routes; deny returns 403* | ☐ |
| 5 | *SQLite state backend* | *—* | <!-- owner --> | 8 | *`StorageBackend` implementation using SQLite; query by resource type/tag* | ☐ |
| 6 | *Fidelity SLOs + Azure baseline tests* | *—* | <!-- owner --> | 6 | *Automated comparison against real Azure responses; fidelity score ≥ 95%* | ☐ |
| 7 | *CLI latency profiling* | *—* | <!-- owner --> | 2 | *`pprof` integration; 90th percentile < 200ms for common commands* | ☐ |
|   | **Phase total** | | | **~32** | | |

**Exit Criteria for Phase N+3:**
- [ ] Zero fuzz-discovered crashes in parsers
- [ ] RBAC enforcement on all simulator routes
- [ ] Azure fidelity score measured and tracked
- [ ] Release process is fully automated end-to-end

---

## 6. Success Metrics & KPIs

> ⚑ GUIDANCE: Define measurable outcomes for each improvement theme.
> Use SMART criteria: Specific, Measurable, Achievable, Relevant, Time-bound.
> Automate measurement where possible (CI reports, dashboards).

### 6.1 Technical Health

| KPI | Current Baseline | Phase N+1 Target | Phase N+2 Target | Phase N+3 Target | How to Measure |
|-----|-----------------|-------------------|-------------------|-------------------|----------------|
| Overall test coverage | *~65%* | *≥ 70%* | *≥ 80%* | *≥ 85%* | `go tool cover` in CI |
| Min per-module coverage | *11.8%* | *≥ 70%* | *≥ 70%* | *≥ 75%* | `go tool cover` per package |
| Largest single file (LOC) | *2,269* | *< 500* | *< 500* | *< 400* | `wc -l` on `store*.go` |
| Known security vulns | *Unknown* | *0 critical/high* | *0 critical/high* | *0 any severity* | `govulncheck` in CI |
| CI pipeline stages | *4* | *6* | *7* | *8+* | Count workflow steps |
| Storage backend options | *1 (JSON file)* | *1* | *3 (mem/hybrid/WAL)* | *4 (+SQLite)* | Feature availability |

### 6.2 Developer Experience

| KPI | Current Baseline | Phase N+1 Target | Phase N+2 Target | How to Measure |
|-----|-----------------|-------------------|-------------------|----------------|
| Time to first working build (new contributor) | *~15 min* | *< 2 min* | *< 2 min* | Time `make dev` from clean clone |
| PR review-to-merge time | *<!-- measure -->* | *< 24h median* | *< 24h median* | GitHub PR analytics |
| Release lead time (tag → binaries published) | *Manual* | *< 10 min (automated)* | *< 10 min* | Workflow run time |
| Open contributor questions (unanswered > 7d) | *<!-- measure -->* | *< 3* | *< 2* | GitHub Issues/Discussions |

### 6.3 Reliability & Fidelity

| KPI | Current Baseline | Phase N+2 Target | Phase N+3 Target | How to Measure |
|-----|-----------------|-------------------|-------------------|----------------|
| Simulator uptime under load (1h stress test) | *Not tested* | *100% (no crashes)* | *100%* | Concurrent stress test suite |
| API fidelity score vs. real Azure | *Not measured* | *≥ 90%* | *≥ 95%* | Automated baseline comparison |
| Contract violations caught pre-merge | *0 (no contracts)* | *100% (all checked)* | *100%* | Contract test CI step |
| Cross-service event delivery rate | *N/A* | *100% (in-process)* | *100%* | Event bus integration test |

### 6.4 Custom KPIs

> ⚑ GUIDANCE: Add project-specific KPIs that matter to your stakeholders.

| KPI | Baseline | Target | Timeframe | How to Measure |
|-----|----------|--------|-----------|----------------|
| <!-- KPI --> | <!-- baseline --> | <!-- target --> | <!-- when --> | <!-- how --> |

---

## 7. Risks & Dependencies

> ⚑ GUIDANCE: Identify what could derail each phase. Assign a likelihood
> (L/M/H), impact (L/M/H), and a mitigation strategy.

### 7.1 Risk Register

| ID | Risk | Likelihood | Impact | Phase Affected | Mitigation |
|----|------|-----------|--------|----------------|------------|
| R-1 | *State store decomposition introduces subtle concurrency bugs* | Medium | High | N+1 | *Extensive testing with `-race` flag; keep existing tests as regression suite; decompose incrementally (one domain at a time)* |
| R-2 | *Storage backend abstraction scope creep* | Medium | Medium | N+2 | *Time-box to 8 days; InMemory + Hybrid first, WAL as stretch goal; defer SQLite to N+3* |
| R-3 | *Event bus adds complexity without immediate user value* | Low | Medium | N+2 | *Gate on having ≥ 2 concrete cross-service use cases before starting* |
| R-4 | *API version validation breaks existing tests/users* | Medium | Medium | N+2 | *Default to permissive mode; strict mode is opt-in; feature flag protected* |
| R-5 | *Team capacity reduced (attrition, competing priorities)* | Medium | High | All | *Prioritize Quick Wins first; each phase item is independently shippable; re-scope at quarterly review* |
| R-6 | *Azure API surface evolves faster than simulator* | High | Medium | N+2, N+3 | *Contract tests detect drift early; focus on stable GA APIs; community contributions for niche services* |
| R-7 | <!-- Risk --> | <!-- L/M/H --> | <!-- L/M/H --> | <!-- Phase --> | <!-- Mitigation --> |

### 7.2 External Dependencies

| Dependency | Type | Owner | Impact if Unavailable | Mitigation |
|-----------|------|-------|----------------------|------------|
| *Go 1.21+ (`log/slog`)* | Runtime | Go team | T-3 blocked | *Already on Go 1.22+ in CI matrix* |
| *GoReleaser* | Build tool | Open source | PR-1 blocked | *Config already exists; well-maintained project* |
| *`govulncheck`* | Security tool | Go team | T-4 degraded (not blocked) | *Can use `nancy` or `trivy` as fallback* |
| *Azure REST API documentation* | Reference | Microsoft | T-7, T-8 slower | *Use OpenAPI specs from `azure-rest-api-specs` GitHub repo* |
| <!-- Dependency --> | <!-- Type --> | <!-- Owner --> | <!-- Impact --> | <!-- Mitigation --> |

### 7.3 Internal Dependencies (Between Improvements)

```
T-1 (decompose store) ──► T-5 (storage backend)
                     ──► T-6 (event bus)

T-2 (coverage gaps) ───► T-8 (contract tests)

T-3 (structured logging) ──► T-9 (request tracing)

T-4 (CI hardening) ───► PR-1 (auto release) ───► PR-3 (changelog automation)
```

---

## 8. Review Cadence

> ⚑ GUIDANCE: This document is a living artifact. Define when and how it gets
> reviewed, who participates, and what decisions are made at each review.

### 8.1 Review Schedule

| Review Type | Frequency | Participants | Duration | Agenda |
|------------|-----------|-------------|----------|--------|
| **Phase Kickoff** | Start of each phase | Full team | 60 min | Review phase scope; assign owners; confirm capacity; identify risks |
| **Progress Check** | Bi-weekly | Tech lead + item owners | 30 min | Status of in-progress items; blocker resolution; scope adjustments |
| **Phase Retrospective** | End of each phase | Full team | 60 min | What shipped; what slipped; lessons learned; update metrics |
| **Quarterly Strategic Review** | Quarterly | Team + stakeholders | 90 min | KPI review; re-prioritize backlog; approve next phase scope; update this document |

### 8.2 Review Checklist

At each quarterly review, answer:

- [ ] Are the **pain points** in §2.2 still accurate? Add new ones, close resolved ones.
- [ ] Have any **bottlenecks** in §2.3 shifted? Update priorities accordingly.
- [ ] Is the **prioritization matrix** in §4 still correct given new information?
- [ ] Did completed items deliver the expected **KPI improvements** in §6?
- [ ] Are there new **risks** in §7 that weren't anticipated?
- [ ] Is the **phase N+2/N+3 scope** still relevant, or does it need re-scoping?
- [ ] Should any items be **promoted** (backlog → next phase) or **demoted** (next phase → backlog)?

### 8.3 Document Maintenance

| Action | When | Who |
|--------|------|-----|
| Update item status (☐ → ☑) | As items complete | Item owner |
| Update KPI baselines | After each measurement | Tech lead |
| Archive completed phases | At phase retrospective | Document owner |
| Add new improvement proposals | Anytime (via PR) | Any team member |
| Full document review & refresh | Quarterly | Document owner |

---

## Appendix A: Decision Log

> ⚑ GUIDANCE: Record key decisions made during reviews. This provides
> an audit trail for why the roadmap evolved the way it did.

| Date | Decision | Rationale | Decided By |
|------|---------|-----------|-----------|
| *<!-- date -->* | *Example: Deferred SQLite backend to Phase N+3* | *Team capacity insufficient for N+2; JSON + WAL covers 90% of use cases* | *<!-- who -->* |
| *<!-- date -->* | *Example: Promoted `make dev` from P2 to P1* | *Three new contributors struggled with setup in the last month* | *<!-- who -->* |
| <!-- date --> | <!-- decision --> | <!-- rationale --> | <!-- who --> |

---

## Appendix B: Blank Templates

### B.1 New Improvement Proposal

Copy and paste into the appropriate category table in §3:

```markdown
| ID | Improvement | Description | Addresses | Effort | Impact |
|----|-------------|-------------|-----------|--------|--------|
| X-# | Title | What, why, and how in 1–2 sentences | PP-#, BN-# | Low/Med/High | Low/Med/High |
```

### B.2 New Risk Entry

Copy and paste into §7.1:

```markdown
| ID | Risk | Likelihood | Impact | Phase Affected | Mitigation |
|----|------|-----------|--------|----------------|------------|
| R-# | Description | L/M/H | L/M/H | N+? | Strategy |
```

### B.3 New KPI

Copy and paste into the appropriate table in §6:

```markdown
| KPI | Baseline | Target | Timeframe | How to Measure |
|-----|----------|--------|-----------|----------------|
| Name | Current value | Goal | By when | Tool / method |
```

---

*Document generated from analysis of azfloci at 17 completed phases, 14,445 LOC, 35 test files.*  
*Cross-references: [IMPROVEMENT_ROADMAP.md](../IMPROVEMENT_ROADMAP.md) | [ROADMAP.md](../ROADMAP.md) | [Architecture](architecture.md)*
