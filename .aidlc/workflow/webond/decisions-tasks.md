# Tasks Decisions (D4)

## Context Summary
WeBond design complete: ~12 FE modules + ~13 BE services, 20 entities, REST API under `/api`, 5 mocked integrations + auto price-generator, 6 PBT properties. Stack: React/TS/Vite/Tailwind + Go/Gin + MySQL, Docker Compose. Solo dev, comprehensive mode. Workshop hands-on policy fixes the execution order: Setup → Frontend on mock JSON → wire Go API + MySQL → verify → deploy (optional). Machine has Node/npm/git but NOT Go/Docker/MySQL yet, so an env install script is part of the Setup wave.

---

## Decision Questions

### D4-1: Task breakdown strategy
**Question**: How to structure the task list?
- 1) **By execution wave matching the hands-on policy** — Wave 0 Setup (incl. env install script), Wave 1 FE+mock (all 9 pages usable), Wave 2 Backend+MySQL+swap, Wave 3 verify/test, Wave 4 deploy(optional) **(Recommended)**
- 2) By architectural layer (all data layer → all services → all UI)
- 3) By feature vertical slice (auth end-to-end, then trading end-to-end, ...)
- 4) Other: _______

**Answer**: By execution wave matching the hands-on policy

---

### D4-2: Implementation approach (tests)
**Question**: When are tests written relative to code?
- 1) **Test-after for UI; property/unit tests alongside financial logic** (money math gets tests as it's written); verify as a later wave per policy **(Recommended)**
- 2) Strict TDD everywhere (tests first)
- 3) Tests entirely deferred to the verify wave
- 4) Other: _______

**Answer**: Strict TDD everywhere

---

### D4-3: Component build priority
**Question**: Which part comes first within the frontend wave?
- 1) **Layout/design-system/tokens + auth shell first, then Market→AssetDetail→Portfolio→Dashboard→DCA→History→Profile** (navigation spine first, then pages) **(Recommended)**
- 2) Dashboard first (highest-value screen)
- 3) Trading flow first (core business)
- 4) Other: _______

**Answer**: Layout/design-system/tokens + auth shell first, then Market→AssetDetail→Portfolio→Dashboard→DCA→History→Profile

---

### D4-4: Integration strategy during dev (mock → API)
**Question**: How to handle the data layer while building?
- 1) **Build FE against the mock-JSON service impl first; add Go API + MySQL, then flip `VITE_DATA_SOURCE=api`; keep mock impl for tests/offline** **(Recommended — matches policy & D3-6)**
- 2) Build FE directly against the API from day one (needs Go+DB up first)
- 3) Use MSW to intercept network instead of a service interface
- 4) Other: _______

**Answer**: Build FE against the mock-JSON service impl first; add Go API + MySQL, then flip `VITE_DATA_SOURCE=api`; keep mock impl for tests/offline
---

### D4-5: Testing scope to include as tasks
**Question**: Which test types get explicit tasks?
- 1) **Backend unit + property (money) + key integration (service+MySQL) + FE formatting/component tests on core flows** (per testing-strategy.md) **(Recommended)**
- 2) Unit tests only
- 3) Full pyramid incl. E2E (Playwright) + load (k6)
- 4) Other: _______

**Answer**: Backend unit + property (money) + key integration (service+MySQL) + FE formatting/component tests on core flows

---

### D4-6: Task granularity
**Question**: How big is one task?
- 1) **Single-concern, ~1–2 days** (e.g., "Build Market page + filters", "TradingService buy/sell MF") **(Recommended for solo)**
- 2) Atomic micro-tasks (per file/function)
- 3) Multi-concern chunks (per whole wave)
- 4) Other: _______

**Answer**: Single-concern, ~1–2 days

---

### D4-7: Execution model
**Question**: Sequential or parallel execution?
- 1) **Sequential by wave, stop-and-verify at the end of each wave** (solo dev; matches policy "หยุดรอผู้ใช้ยืนยันเมื่อจบแต่ละกลุ่ม") **(Recommended)**
- 2) Parallel waves with sync points
- 3) Other: _______

**Answer**: Sequential by wave, stop-and-verify at the end of each wave

---

### D4-8: Environment setup & install script (Wave 0)
**Question**: Given Go/Docker/MySQL are not installed, what does the Setup wave + env install script do?
- 1) **Scaffold everything to run + provide a `scripts/setup-env.sh` that checks and installs missing tools via Homebrew (Go, Docker, optional MySQL client) and verifies versions; FE runnable in mock mode on Node immediately; full-stack run after script + Docker Desktop launch. Includes README run + demo credentials** **(Recommended)**
- 2) Script that only checks/prints what's missing (no auto-install)
- 3) No script — document manual install steps only
- 4) Other: _______

**Answer**: setup for realbuild, in development phase use npm run dev.

---

## Decisions Summary
<!-- Populated after you answer. One line per decision. -->
- D4-1 Breakdown: By execution wave (W0 Setup → W1 FE+mock → W2 Backend+MySQL+swap → W3 verify → W4 deploy optional)
- D4-2 Test approach: TDD within each wave for critical logic (money/trading/DCA/wallet, Go) — tests first; test-after for UI; verify at end of each wave (C1 resolved → option 1)
- D4-3 Build priority: Layout/design-system/tokens + auth shell first, then Market→AssetDetail→Portfolio→Dashboard→DCA→History→Profile
- D4-4 Integration strategy: FE on mock-JSON impl first; add Go API + MySQL; flip VITE_DATA_SOURCE=api; keep mock for tests/offline
- D4-5 Test scope: Backend unit + property (money) + key integration (service+MySQL) + FE formatting/component tests on core flows
- D4-6 Granularity: Single-concern, ~1–2 days
- D4-7 Execution model: Sequential by wave, stop-and-verify at end of each wave
- D4-8 Setup/install script: Env install script + Docker/MySQL are for the REAL build; during development use `npm run dev` (frontend mock mode). Script lives in Wave 0 but full-stack install is only needed for the real/backend build.

---

**Instructions**: Fill in your answers above and respond with "done" — or say "use recommendations" to auto-fill all recommended options.

---

## Validation Notes
- **C1 (D4-2) resolved → option 1**: TDD applied per wave for critical backend logic (money math, trading, DCA, wallet) — write tests first there; UI uses test-after; each wave ends with a verify/stop checkpoint. Reconciles with hands-on policy (UI-first, working software) and D4-5 test scope.
- D4-8: env install script + Docker/MySQL are for the real/backend build; development uses `npm run dev` (frontend mock mode). No conflict — script is created in Wave 0 but full-stack install is only required when building the real backend (Wave 2+).
- Checked: DB+sqlc in D3 → migration tasks included (Wave 0/2); testing frameworks in D3 → test tasks included (per-wave + Wave 3); solo + sequential-by-wave → task-count/capacity OK; no cloud deploy in phase 1 → infra/CI tasks deferred to optional Wave 4.
