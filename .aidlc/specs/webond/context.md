# Context Assessment

## Summary
<!-- 10-line max. Downstream phases read ONLY this section. -->
- **Type**: Greenfield
- **Scope**: new
- **Stack**: React 18 + TypeScript (Vite) + Tailwind CSS frontend; Go (Gin) REST API; MySQL 8.0 real DB (sample data seeded into DB). Frontend built first on mock JSON, then wired to the Go API.
- **Architecture**: SPA frontend (layered: pages → components → hooks → services) calling a layered Go REST API (handler → service → repository → MySQL).
- **Feature**: WeBond — a retail investment platform to buy/sell and manage Bonds, Mutual Funds and Gold in one portfolio, with DCA, history and an insight dashboard.
- **Impact**: New standalone
- **Complexity**: High — 26 FRs across 8 domains (Account, Catalog, Wallet, Trading, DCA, History, Portfolio, Dashboard), 3 user types, 5 mocked external systems.
- **Recommendations**: Personas Yes, Units No (comprehensive mode per workshop policy), NFR Yes

## Project Overview
- **Type**: Greenfield
- **Assessment Date**: 2026-10-28

## Technology Stack
- **Languages**: TypeScript (frontend), Go (backend — later phase)
- **Frameworks**: React 18 + Vite + Tailwind CSS + React Router + TanStack Query + Recharts; Gin (backend)
- **Build System**: npm / Vite (frontend); Go modules (backend)
- **Testing**: Vitest + React Testing Library (frontend); Go testing + testify (backend)
- **Infrastructure**: Phase 1 local (Vite dev server, mock JSON). Target: Docker Compose, AWS ECS Fargate + RDS MySQL + ElastiCache Redis + S3/CloudFront.

## Patterns & Conventions
N/A — greenfield project. Conventions to be set during design (D3). Baseline from steering: camelCase vars/functions, PascalCase components/classes, UPPER_SNAKE_CASE constants, one component per file, interfaces over types for public APIs, explicit return types on exported functions, no `any`.

## Codebase Analysis
N/A — greenfield project (no existing source).

## Feature Impact

**Affected Areas**: New standalone

| Area | Impact | Reason |
|------|--------|--------|
| Frontend SPA | New | All 9 UI pages built from scratch |
| Mock data layer | New | Phase-1 data source (JSON) feeding all screens |
| Backend REST API | New (later phase) | Go/Gin service per FR-26, phase 1 mocked |
| Database schema | New (later phase) | MySQL `webond` schema (users, wallets, assets, orders, holdings, dca_plans, transactions) |

## Recommendations

- Story Count: High (26 functional requirements + 16 exceptions + 17 business rules)
- Domain Boundaries: Account/Auth, Asset Catalog (Market), Wallet, Trading (MF/Bond/Gold), DCA, History, Portfolio, Dashboard, Notification, Admin
- User Types: Investor (registered), Guest, Admin
- Integration Points: Mock Bank/Payment, Fund House, Bond Market, Gold Price, Notification (all mocked phase 1)
- **Personas**: Yes — distinct user classes (Investor, Guest, Admin) with different goals and access
- **Units**: No — workshop policy mandates comprehensive mode (no unit decomposition) for hands-on projects
- **NFR**: Yes — security (bcrypt, JWT, OWASP, PDPA), financial accuracy (DECIMAL), performance, responsive UI are explicit and material

## Scope

- **Detected scope**: new
- **Rationale**: Greenfield workspace (no source code), requirements document explicitly states "Greenfield (new system – no existing source code / database)".
- **Phases skipped**: Decomposition is bypassed per workshop hands-on policy (comprehensive mode). Full path otherwise.

## Recommended Workflow

```
┌─────────────┐
│   Context   │ ✅ (this phase)
└──────┬──────┘
       ▼
┌─────────────┐
│ Requirements│  user stories + EARS, personas
└──────┬──────┘
       ▼
┌─────────────┐
│   Design    │  D3 tech decisions, architecture, data model, API, UI
└──────┬──────┘
       ▼
┌─────────────┐
│    Tasks    │  sequenced waves (Setup → FE+mock → FE+backend → verify)
└──────┬──────┘
       ▼
┌─────────────┐
│  Implement  │  code generation, runnable web app
└──────┬──────┘
       ▼
┌─────────────┐
│ Build & Test│  compile, run dev server, verify flows, tests
└──────┬──────┘
       ▼
┌─────────────┐
│   Deploy    │  optional (last)
└─────────────┘
```
(Comprehensive mode — no unit decomposition.)

## External References

| Source | Type | What was used |
|--------|------|---------------|
| initial-requirements/project-requirements/(MSC) AIDLC Workshop - Usecase WeBond.md | Requirements spec | FRs, NFRs, exceptions, business rules, tech stack, scope |
| initial-requirements/project-requirements/images/01..09 *.png | Wireframes | UI layout, navigation, page content for all 9 screens |
