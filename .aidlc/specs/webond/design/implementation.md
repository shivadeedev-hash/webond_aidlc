# Implementation

## Code Organization
- **Pattern**: Layered on both tiers. Monorepo (single git repo).
- **Repository type**: single repo with `web/` (frontend), `api/` (backend), `docker-compose.yml` at root.

### Directory structure
```
WeBond/
├── docker-compose.yml          # web + api + mysql
├── .gitignore
├── README.md                   # run instructions + demo credentials
├── .env.example                # root env sample
├── web/                        # React 18 + TS + Vite + Tailwind
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── index.html
│   ├── .env.example            # VITE_DATA_SOURCE=mock|api, VITE_API_URL
│   ├── public/
│   └── src/
│       ├── main.tsx
│       ├── App.tsx             # router + QueryClient + layout
│       ├── pages/              # Login,Register,ForgotPassword,ResetPassword,Dashboard,Market,AssetDetail,Portfolio,DCA,History,TransactionDetail,Profile
│       ├── components/{layout,market,trade,dca,portfolio,dashboard,history,wallet,risk,charts,ui}/
│       ├── hooks/              # useAuth + per-domain query hooks
│       ├── services/{auth,asset,wallet,order,dca,portfolio,dashboard,history,notification}/{index.ts,mock.ts,http.ts}
│       ├── mocks/              # *.json seed used by mock.ts (interim)
│       ├── types/
│       ├── lib/{format,i18n,apiClient,errors,constants}.ts
│       ├── styles/tokens.css   # design tokens (CSS variables)
│       └── test/               # Vitest + RTL
├── api/                        # Go 1.23 + Gin
│   ├── go.mod
│   ├── cmd/webond/main.go      # entry: config, DB, router, graceful shutdown
│   ├── sqlc.yaml
│   ├── db/
│   │   ├── migrations/         # golang-migrate *.up.sql / *.down.sql
│   │   ├── queries/            # *.sql for sqlc
│   │   └── seed/seed.sql       # rich sample data + demo investor
│   └── internal/
│       ├── config/
│       ├── http/{router.go,response.go,middleware/}
│       ├── service/            # auth, wallet, catalog, trading, dca, portfolio, dashboard, history, bondlifecycle, notification, admin, pricegen
│       ├── repo/               # repository interfaces + sqlc impl
│       ├── db/                 # sqlc-generated code
│       ├── money/              # decimal helpers (shopspring)
│       └── domain/             # entities, error codes
└── .aidlc/ , .kiro/            # workflow + skills (not shipped)
```

### Naming conventions
- Frontend: camelCase vars/functions, PascalCase components/classes, UPPER_SNAKE_CASE constants, one component per file, `index.ts` barrels, interfaces over types for public APIs, explicit return types on exports, no `any`.
- Go: idiomatic Go (MixedCaps), package-per-concern, errors wrapped with `%w`, exported types documented.

## Technology Stack (resolved versions)
> Versions pinned from training knowledge; marked ⚠️ unverified where live registry lookup didn't return an exact patch. Confirm/adjust at `npm install` / `go mod tidy`.

| Layer | Dependency | Version |
|-------|-----------|---------|
| Frontend runtime | Node.js | 24.x (local: 24.14.1 detected) |
| | React / React DOM | 18.3.x ⚠️ |
| | TypeScript | 5.6.x ⚠️ |
| | Vite | 5.4.x ⚠️ |
| | Tailwind CSS | 3.4.x ⚠️ |
| | React Router | 6.x ⚠️ |
| | TanStack Query | 5.x ⚠️ |
| | Recharts | 2.x ⚠️ |
| | Vitest + React Testing Library | 2.x / 16.x ⚠️ |
| Backend | Go | 1.23.x ⚠️ |
| | Gin (gin-gonic/gin) | 1.10.x ⚠️ |
| | golang-jwt/jwt | v5 ⚠️ |
| | shopspring/decimal | 1.4.x ⚠️ |
| | xuri/excelize | v2 ⚠️ |
| | sqlc | 1.27.x ⚠️ |
| | golang-migrate | 4.x ⚠️ |
| | go-sql-driver/mysql | 1.8.x ⚠️ |
| | stretchr/testify | 1.9.x ⚠️ |
| | leanovate/gopter (property tests) | 0.2.x ⚠️ |
| Database | MySQL | 8.0 |
| Infra | Docker Compose | v2 |

## Development Setup

### Prerequisites (IMPORTANT — current machine status)
- **Node.js 20+** — ✅ present (24.14.1)
- **npm** — ✅ present (11.11.0)
- **Go 1.22+** — ❌ NOT installed → required for backend
- **Docker + Docker Compose** — ❌ NOT installed → required for the one-command MySQL + full-stack run
- **MySQL 8.0** — ❌ no local client/server → provided via Docker (preferred) or install locally

> See operations.md § Local Environment Setup for exact install steps. Phase-1 "Setup" task will scaffold the repo; the **frontend (mock mode) runs on Node alone** without Go/Docker/MySQL, so the UI is demonstrable immediately; the full stack needs Go + Docker.

### Setup commands
```
# Frontend only (mock mode) — works now with Node:
cd web && npm install && npm run dev        # http://localhost:5173, VITE_DATA_SOURCE=mock

# Full stack (after installing Docker + Go):
docker compose up --build                   # web:5173/8080, api:8080, mysql:3306, auto-migrate + seed
# or run API on host:
cd api && go mod tidy && go run ./cmd/webond
```

### Environment variables
| Variable | Description | Example |
|----------|-------------|---------|
| `VITE_DATA_SOURCE` | frontend data layer | `mock` or `api` |
| `VITE_API_URL` | API base URL | `http://localhost:8080/api` |
| `APP_ENV` | backend env | `development` |
| `PORT` | API port | `8080` |
| `DATABASE_URL` | MySQL DSN | `webond:webond@tcp(mysql:3306)/webond?parseTime=true&loc=Asia%2FBangkok` |
| `JWT_ACCESS_SECRET` | access token secret | `change-me-dev` |
| `JWT_REFRESH_SECRET` | refresh token secret | `change-me-dev` |
| `LOG_LEVEL` | log level | `info` |
| `SHUTDOWN_TIMEOUT_MS` | graceful shutdown | `30000` |

> Secrets only in env/.env (gitignored). `.env.example` holds safe placeholders. No secrets in source (NFR-01, 12-factor).

## Testing
- **Unit (frontend)**: Vitest + React Testing Library — `npm run test`.
- **Unit (backend)**: `go test ./...` with testify.
- **Property-based (backend)**: gopter for money math (see correctness.md) — part of `go test ./internal/money/... ./internal/service/...`.
- **Integration (backend)**: service+repo against a MySQL test DB (Docker) — `go test -tags=integration ./...`.
- **Coverage target**: ≥ 70% for backend business logic (NFR-08).
