# Verancial

Verancial is a personal finance application that ingests bank statement
exports (CSV files from supported banks), normalizes and de-duplicates the
transactions, auto-categorizes them, and stores them per user. From there it
can generate export reports for third-party budgeting apps — currently
[BudgetBakers](https://budgetbakers.com/) — so a user's bank transactions
show up in their budgeting tool of choice.

Supported banks today (see `shared/constants`): Scotiabank (chequing and
credit card), Nubank, Wise, and First Tech Federal Credit Union (chequing and
credit card).

## Architecture

The system is a Go monorepo with one React frontend, three Go backend
services, a shared Go module, a Postgres/Supabase database, a Redis-backed
task queue, and a couple of standalone Python/Node automation scripts.

### Services

- **`api/`** — the main REST API (Gin), used by the frontend. Routes live
  under `/api/v0`:
  - `POST /report/process` — kicks off transaction-report processing
  - `POST /app-integration/generate` — kicks off a BudgetBakers export
  - `GET /dashboard/user` — dashboard stats for the logged-in user
  - `GET /transaction/list/:bankId` — list transactions for a bank
  - `GET /bank/:bankId` — bank account stats
  All routes are behind JWT auth middleware (`shared/auth`).

- **`login/`** — authentication service (Gin), routes under `/login/v0`:
  - `POST /user/signin`, `POST /user/signup` (public)
  - `DELETE /admin/delete`, `PUT /admin/update` (JWT-protected)
  Issues/validates JWTs, hashes passwords with bcrypt.

- **`data-process-worker/`** — parses uploaded bank CSV exports into
  `models.Transaction` rows. Each bank has its own parser under
  `pkg/models/<bank>` and `pkg/report/<bank>`. It de-duplicates using a
  transaction fingerprint and runs transactions through a rule-based
  category guesser (`pkg/category-guesser`, backed by
  `pre_defined_categories.json`) before bulk-inserting into Postgres.

- **`app-integration-worker/`** — generates and submits export reports to
  external finance apps. Currently implements a BudgetBakers generator
  (`pkg/generators/budgetbakers`) that builds a CSV report from a user's
  transactions for upload/import into BudgetBakers.

Both workers can run in two modes, controlled by an `AsyncProcessing` flag:
- **Sync mode**: exposed as a small Gin HTTP endpoint
  (`/dpw/v0/process_report`, `/aiw/v0/process_app_report`) that runs the job
  inline.
- **Async mode** (the queue-based path): an [asynq](https://github.com/hibiken/asynq)
  worker listening on Redis queues (`critical`/`default`/`low`).

The `api` service enqueues work onto Redis via `shared/task` (`AsyncQueue`),
using task patterns defined in `shared/types` (`PatternReportProcess`,
`PatternAppIntegration`). The two workers consume those tasks asynchronously
— this is the primary inter-service communication mechanism, not synchronous
REST calls between services.

### Shared module (`shared/`)

Common Go code imported by all four services:
- `auth` — JWT creation/validation
- `constants` — bank IDs and other shared constants
- `errors` — Gin error-handling helpers (`ErrorRoute`)
- `flags` — env-based configuration loading (`VERANCIAL_DEPLOY_ENV`-driven)
- `models` — `User`, `Transaction`, `BankAccount`, `BankCredentials`,
  `FinancialAppCredentials`, `BudgetBakers` (export row), dashboard DTOs
- `repository` — data-access interface, with a Postgres implementation in
  `repository/postgresRepository` (users, transactions, bank accounts)
- `scripts` — Mage-based DB migration scripts (`mage.go`) and local dev
  helper scripts
- `task` — the asynq-based task/queue client (`Task` interface,
  `AsyncQueue`)
- `types` — queue payload types and task pattern names
- `validator` — shared request validation

### Database

Postgres, hosted on [Supabase](https://supabase.com/). SQL migrations live
in `migrations/`:
- `run_migrations.sql` — creates `users` and `transactions` tables, with
  Row Level Security policies (per-user access plus a service-role bypass)
- `002_create_bank_accounts.sql` — creates `bank_accounts` (supported banks)
  and `user_bank_accounts` (per-user bank connections), also RLS-protected

Local Postgres via docker-compose has been removed in favor of Supabase;
the commented-out block in `docker-compose.yaml` shows how to bring a local
Postgres container back if needed.

### Queue

Redis (`worker-redis` in docker-compose, port 6379) backs the asynq task
queue used for async report processing and app-integration jobs.

### Frontend

`frontend/` — React 18 + TypeScript, bootstrapped with Create React App via
CRACO. Key pieces under `frontend/src`:
- `components/Login.tsx`, `Dashboard.tsx`, `Bank.tsx`
- `services/api.ts` — Axios client for the `api` service
Served on port 3000; talks to the `api` (8080) and `login` (8081) services.

### Python / Node automation scripts

`python-scripts/` (not part of the Docker stack, run manually):
- `get_last_transaction_budgetbakers.py` — Selenium script that logs into
  BudgetBakers' web UI and reads back the last imported transaction date, so
  the app-integration worker knows where to resume exporting from
- `convertBRLToAnyCurrency_Budget.js` — currency conversion helper for BRL
- `requirements.txt` — Python deps (Selenium, webdriver-manager, etc.)

## Running locally

Requirements: Docker, Go 1.18+, Node (for frontend dev), and a Supabase
Postgres instance (connection configured via env vars consumed by
`shared/flags`).

### Full stack via Docker Compose

```
make all          # docker-compose up
make all_build     # docker-compose up --build
```

This brings up `api` (:8080), `login` (:8081), `frontend` (:3000), and
`worker-redis` (:6379). `data-process-worker` and `app-integration-worker`
are not in `docker-compose.yaml` — run them locally (see below) or add
services for them as needed.

### Running services individually (Go, local)

```
make start_api_local                       # api on :8080
make start_login_local                     # login on :8081
make start_report_process_worker_local     # data-process-worker
make start_app_integration_worker_local    # app-integration-worker
make start_frontend_local                  # frontend via npm run dev
make start_redis                           # just Redis, via docker-compose
```

Each `*_local` target sets `VERANCIAL_DEPLOY_ENV=local` and runs `go run
main.go` directly from the service directory, picking up local env config
files via `shared/flags`.

### Database migrations

```
make migrate_db
```

Runs the Mage-based migration scripts in `shared/scripts` against the
configured (local) database: `migrateUserModel` and
`migrateTransactionModel`. For Supabase, you can alternatively run the SQL
in `migrations/run_migrations.sql` and `migrations/002_create_bank_accounts.sql`
directly in the Supabase SQL editor.

## Directory structure

```
api/                      REST API service (Gin) — frontend-facing
  cmd/server/              entrypoint wiring (flags, builder)
  pkg/builder/             dependency/config builder
  pkg/handlers/            route handlers (dashboard, bank, report, app-integration)
  pkg/handlers/transaction/ transaction listing handler
  pkg/middlewares/         auth middleware
  pkg/validator/           request validation

login/                    Auth service (Gin)
  pkg/handlers/             sign-up/sign-in/update/delete
  pkg/middlewares/          auth middleware
  pkg/constants/, pkg/validator/

data-process-worker/      Bank CSV ingestion + categorization (sync HTTP or asynq worker)
  pkg/models/<bank>/        per-bank CSV row models
  pkg/report/<bank>/        per-bank CSV parsing/processing logic
  pkg/category-guesser/     rule-based transaction categorization
  pkg/handlers/, pkg/helper/, pkg/builder/

app-integration-worker/   Export-to-third-party-app worker (sync HTTP or asynq worker)
  pkg/generators/budgetbakers/  BudgetBakers CSV report generator
  pkg/generators/helper/
  pkg/handlers/, pkg/types/, pkg/builder/

shared/                   Shared Go module used by all services
  auth/, constants/, errors/, flags/, models/, types/, validator/
  repository/                data-access interface + postgresRepository implementation
  task/                       asynq queue client

frontend/                 React + TypeScript SPA (CRACO/CRA)
  src/components/           Login, Dashboard, Bank views
  src/services/             API client (Axios)

migrations/                Supabase/Postgres SQL migrations (users, transactions,
                            bank_accounts, user_bank_accounts + RLS policies)

python-scripts/            Standalone Selenium/Node helper scripts (BudgetBakers
                            last-transaction lookup, BRL currency conversion)

docker-compose.yaml         api, login, frontend, worker-redis
makefile                    local dev / docker / migration targets
```

## Contributing

1. Open an issue to discuss your changes.
2. Fork the repository.
3. Create a new branch for your changes.
4. Make your changes and commit them.
5. Push your changes to your forked repository.
6. Open a pull request to the main repository.

## Contact

`verancial@verasthiago.com` or https://www.linkedin.com/in/verasthiago/.
