# Structure

## Summary
- Single repo, greenfield. Frontend under `web/` (React SPA); backend under `api/` (Go/Gin); MySQL via Docker Compose.
- `web/src/mocks/` holds temporary mock JSON used only during the frontend-first step, then the data layer is switched to the Go API; sample data is seeded into MySQL.
- Entry points: `web/src/main.tsx` → `App.tsx` (SPA); `api/cmd/.../main.go` (REST API).

## Repository
- **Type**: Single repo (monorepo-ready: `web/`, later `api/`). Pending D3 confirmation.
- **Root**: project root with `.gitignore`, README, and `web/` application directory.

## Key Directories

| Directory | Purpose | Key contents |
|-----------|---------|--------------|
| `web/` | React SPA | app source, config |
| `web/src/pages/` | Route-level pages | Login, Register, Dashboard, Market, AssetDetail, Portfolio, DCA, History, TransactionDetail, Profile |
| `web/src/components/` | Reusable UI components | layout (Header/Nav), cards, tables, charts, forms, dialogs |
| `web/src/hooks/` | Data + state hooks | TanStack Query hooks over the data layer |
| `web/src/services/` | Data layer (interface) | asset, wallet, order, dca, history, portfolio, auth services |
| `web/src/mocks/` | Temporary mock JSON (frontend-first step) | funds, bonds, gold, prices, holdings, orders, transactions, dca plans, wallet — later replaced by Go API calls |
| `web/src/types/` | Shared TypeScript types | domain models, DTOs |
| `web/src/lib/` | Utilities | formatting (THB, units, dates), constants, API client |
| `api/` | Go (Gin) REST API | cmd entry, handlers, services, repositories, models |
| `api/migrations/` | MySQL schema migrations | golang-migrate SQL files (users, wallets, funds, bonds, gold, prices, orders, holdings, dca_plans, transactions, ...) |
| `api/seed/` | Sample data seed | seeds the same sample data into MySQL (source for mock external APIs per FR-26) |
| (root) `docker-compose.yml` | Local orchestration | web, api, mysql |

> Greenfield: directory layout above is the recommended target; design phase (D3) confirms and may refine.

## Key Files

| File | Purpose | Notes |
|------|---------|-------|
| `web/package.json` | Frontend deps & scripts | Pending creation at setup |
| `web/vite.config.ts` | Vite config | Pending creation |
| `web/tailwind.config.js` | Tailwind + design tokens | Maps design-system tokens |
| `web/src/main.tsx` | App bootstrap | Pending creation |
| `web/src/App.tsx` | Router + layout | Pending creation |
| `.gitignore` | Ignore node_modules, build, env | Created at setup (`git init -b main`) |

## Entry Points

| Entry | Type | Description |
|-------|------|-------------|
| `web/src/main.tsx` | SPA bootstrap | Mounts React app, Router, Query client |

(Module dependencies, data flow, key abstractions, test organization, and build/deploy details are defined during the design phase — greenfield.)
