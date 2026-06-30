# Migrations & Schema Audit

Date: 2026-06-30
Scope: `migrations/` SQL files, cross-referenced against `shared/models` (Go structs) and
`shared/repository/postgresRepository` (GORM queries) and `api/pkg/handlers`.

This is a **local, code-only review**. No live Supabase connection was available; nothing
in this audit was executed against a database. New migration files are added under
`migrations/` for later review/application by someone with access to the live project.

---

## 1. Tables defined across migrations

### `migrations/run_migrations.sql` ("Migration 001" + "Migration 002" — see Finding F1 on numbering)

**`users`**
| Column | Type | Notes |
|---|---|---|
| id | VARCHAR(255) PK | default `uuid_generate_v4()::text` |
| created_at | TIMESTAMPTZ | default now() |
| updated_at | TIMESTAMPTZ | default now() |
| deleted_at | TIMESTAMPTZ | soft delete |
| name | VARCHAR(255) NOT NULL | |
| email | VARCHAR(255) UNIQUE NOT NULL | |
| password | VARCHAR(255) NOT NULL | hashed |
| is_admin | BOOLEAN default false | |
| is_verified | BOOLEAN default false | |
| financial_app_credentials | JSONB | |
| bank_credentials | JSONB | |

Indexes: `idx_users_email`, `idx_users_deleted_at`.
RLS: enabled + `FORCE ROW LEVEL SECURITY`. Policies: select/update/delete own (by `sub` claim), insert open (registration), `service_role_users` ALL bypass.

**`transactions`**
| Column | Type | Notes |
|---|---|---|
| id | VARCHAR(255) PK | default `uuid_generate_v4()::text` |
| created_at / updated_at / deleted_at | TIMESTAMPTZ | |
| user_id | VARCHAR(255) NOT NULL | FK → `users(id)` ON DELETE CASCADE |
| date | TIMESTAMPTZ NOT NULL | |
| amount | REAL NOT NULL | |
| payee | VARCHAR(255) | |
| description | TEXT | |
| category | VARCHAR(255) | |
| currency | VARCHAR(10) | |
| bank_id | VARCHAR(255) | **no FK constraint** (see F4) |
| metadata | JSONB | |

Indexes: `idx_transactions_user_id`, `idx_transactions_date`, `idx_transactions_bank_id`, `idx_transactions_user_bank` (user_id, bank_id), `idx_transactions_user_bank_date` (user_id, bank_id, date), `idx_transactions_deleted_at`.
RLS: enabled + FORCE. Policies: select/insert/update/delete own (by `sub` claim), `service_role_transactions` ALL bypass.

### `migrations/002_create_bank_accounts.sql` (also labeled "Migration 002" — duplicate number, see F1)

**`bank_accounts`** (reference/lookup table of supported banks)
| Column | Type | Notes |
|---|---|---|
| id | VARCHAR(255) PK | e.g. `'scotiabank'`, `'nubank'` — **not a UUID** (see F4) |
| name | VARCHAR(255) NOT NULL | |
| display_name | VARCHAR(255) NOT NULL | |
| country_code | VARCHAR(3) | |
| currency | VARCHAR(10) | |
| created_at / updated_at | TIMESTAMPTZ | |
| is_active | BOOLEAN default true | |

Indexes: `idx_bank_accounts_active`.
RLS: enabled. Policy: `bank_accounts_select_active` (SELECT where `is_active = true`, open to all — appropriate for a public reference table). No INSERT/UPDATE/DELETE policy and no service-role bypass policy (see F2).

**`user_bank_accounts`** (join table: user ↔ supported bank)
| Column | Type | Notes |
|---|---|---|
| id | VARCHAR(255) PK | default `uuid_generate_v4()::text` |
| user_id | VARCHAR(255) NOT NULL | FK → `users(id)` ON DELETE CASCADE |
| bank_id | VARCHAR(255) NOT NULL | FK → `bank_accounts(id)` ON DELETE CASCADE |
| created_at / updated_at | TIMESTAMPTZ | |
| is_active | BOOLEAN default true | |
| last_sync_date | TIMESTAMPTZ | |

Unique constraint: `(user_id, bank_id)`.
Indexes: `idx_user_bank_accounts_user_id`, `idx_user_bank_accounts_bank_id`, `idx_user_bank_accounts_active`.
RLS: enabled. Policies: select/insert/update/delete own (by `sub` claim). **No service-role bypass policy** (see F2) — `FORCE ROW LEVEL SECURITY` was also never applied to this table, unlike `users`/`transactions`.

### Relationships (FK graph)
```
users (1) ──< transactions (user_id)
users (1) ──< user_bank_accounts (user_id)
bank_accounts (1) ──< user_bank_accounts (bank_id)
transactions.bank_id  -->  (no FK)  bank_accounts.id   [orphaned reference, see F4]
```

---

## 2. Row Level Security review

| Table | RLS enabled | FORCE RLS | Policies present | Verdict |
|---|---|---|---|---|
| `users` | Yes | Yes | select/update/delete own, insert (open), service_role bypass | OK |
| `transactions` | Yes | Yes | select/insert/update/delete own, service_role bypass | OK |
| `bank_accounts` | Yes | No | select (active only) | OK — read-only reference data, acceptable without write policies since app writes via migration/seed, not user requests |
| `user_bank_accounts` | Yes | **No** | select/insert/update/delete own | **Finding F2 (Medium)** |

### Finding F2 — `user_bank_accounts` has no `FORCE ROW LEVEL SECURITY` and no service-role bypass policy
**Severity:** Medium
**Description:** `users` and `transactions` both explicitly `FORCE ROW LEVEL SECURITY` and define a `service_role_*` ALL policy so that backend/service-role connections can operate without being silently blocked by RLS (table owners bypass RLS by default unless FORCEd, but the *intent* in this codebase — established by the other two tables — is to always add an explicit service-role policy). `user_bank_accounts` links sensitive data (which banks a user has connected) but has neither `FORCE ROW LEVEL SECURITY` nor a `service_role_user_bank_accounts` policy. This is inconsistent with the pattern used elsewhere and should be fixed for parity/predictability, especially since the Go backend (`GetUserBankAccounts`, `GetUserDashboardStats`, `GetBankAccountById`) reads this table directly via GORM with the service/admin DB role and relies on RLS not interfering.
**Recommended migration:** see `migrations/004_user_bank_accounts_rls_hardening.sql` (added in this PR).

### Finding F3 — `bank_accounts` has no explicit service-role write policy
**Severity:** Low
**Description:** `bank_accounts` is seeded via the migration itself (superuser context, bypasses RLS), but if the application ever needs to insert/update bank metadata at runtime via the service role, there is no policy permitting it (only `SELECT WHERE is_active = true` exists). This is currently low-impact since the table is effectively static reference data, but flagging for awareness. Recommend adding a service-role ALL policy for consistency, bundled into the same migration as F2.

---

## 3. Migration numbering / ordering consistency

### Finding F1 — Duplicate migration number "002" / no real sequential migration files
**Severity:** Medium
**Description:**
- `migrations/run_migrations.sql` is a single monolithic script containing inline comments `-- Migration 001: Create Users table` and `-- Migration 002: Create Transactions table`.
- `migrations/002_create_bank_accounts.sql` is a *separate file* also numbered **002**, creating the `bank_accounts` and `user_bank_accounts` tables (which is actually a third migration in sequence, conceptually "003").
- There is no `001_*.sql` file on disk — migration 001 only exists as a section inside `run_migrations.sql`.
- This means the numbering scheme is not actually consistent or self-describing from the filenames alone, and a future contributor adding `003_*.sql` could be misled about ordering, or could collide if they (reasonably) assume `002_create_bank_accounts.sql` was the second migration and add `002_xxx.sql` again.

**Recommendation:** Treat `run_migrations.sql` as migrations 001 (users) + 002 (transactions) conceptually, and rename/renumber `002_create_bank_accounts.sql` to `003_create_bank_accounts.sql` to remove the collision. Because renaming an already-applied migration file is risky (file name often isn't what tracks applied state, but it avoids future confusion), this audit does **not** rename the existing file (to avoid implying the file should be re-run / to avoid breaking any external reference to the existing filename), but new migrations added by this audit are numbered starting at **004** to avoid colliding with the existing `002_*.sql`. A follow-up migration-tooling cleanup is recommended separately (e.g., split `run_migrations.sql` into individual `001_create_users.sql` / `002_create_transactions.sql` files and rename `002_create_bank_accounts.sql` → `003_create_bank_accounts.sql`).

### Finding F1b — `run_migrations.sql` does not include/run the bank_accounts migration
**Severity:** Medium
**Description:** Despite its name, `run_migrations.sql` does **not** `\i` or otherwise include `002_create_bank_accounts.sql`. It only contains the SQL for users + transactions. A developer running "the migration script" (singular, per its own description as "Run this script in your Supabase SQL editor or via CLI") would not get the `bank_accounts` / `user_bank_accounts` tables, and would need to separately know to also run `002_create_bank_accounts.sql`. There is no top-level orchestration (no `001_*.sql`, `002_*.sql`, ... run in a loop by a script, and no Supabase CLI `supabase/migrations` directory convention in use either).
**Recommendation:** Either (a) convert to the standard Supabase CLI convention (`supabase/migrations/<timestamp>_<name>.sql`, each applied in order automatically), or (b) make `run_migrations.sql` a thin orchestrator that `\i`-includes each numbered file in order. Out of scope to implement here since it requires a decision on migration tooling, but flagged as a process risk.

---

## 4. Missing indexes (cross-referenced against repository queries)

Reviewed every query in `shared/repository/postgresRepository/transaction.go`, `bank_account.go`, `user.go`.

| Query | Filter columns | Index present? |
|---|---|---|
| `GetUserByEmail` | `email` | Yes (`idx_users_email`) |
| `GetUserByID` | `id` (PK) | Yes (PK) |
| `GetLastTransactionFromUserBank` | `user_id, bank_id` + `ORDER BY date` | Yes (`idx_transactions_user_bank_date` covers this) |
| `GetAllTransactionsFromUserBankAfterDate` | `user_id, bank_id, date >` | Yes (`idx_transactions_user_bank_date`) |
| `GetTransactions` (paginated list) | `user_id, bank_id` + optional `category IS NULL OR category = ''` + `ORDER BY date DESC` | Base filter covered by `idx_transactions_user_bank_date`; **the `uncategorized` filter branch has no supporting index** |
| `GetTransactionCountFromUserBank` | `user_id, bank_id` | Yes |
| `GetUserBankAccounts` | `user_id, is_active` | **Partial** — `idx_user_bank_accounts_user_id` exists but no composite `(user_id, is_active)`, and no index covers `is_active` filtering together with `user_id` efficiently for larger datasets |
| `GetBankAccountById` | `bank_id, user_id, is_active` | **No composite index** — only single-column indexes exist on `user_bank_accounts` |

### Finding F5 — Missing composite index on `user_bank_accounts(user_id, is_active)` and `(bank_id, user_id, is_active)`
**Severity:** Low (table is small/low cardinality today, but flagged as the table grows and as a query-plan correctness habit)
**Description:** `GetUserBankAccounts` and `GetBankAccountById` both filter on `user_id` + `is_active` (and `bank_id` for the latter). Only single-column indexes exist (`idx_user_bank_accounts_user_id`, `idx_user_bank_accounts_bank_id`, `idx_user_bank_accounts_active`), which Postgres can combine via bitmap-and but a composite index is more efficient for these specific hot-path lookups (dashboard stats, per-bank lookups are called per-request).
**Recommended migration:** see `migrations/004_user_bank_accounts_rls_hardening.sql` includes the RLS fix; index added separately in `migrations/005_add_missing_indexes.sql`.

### Finding F6 — No index supporting the "uncategorized transactions" filter
**Severity:** Low
**Description:** `GetTransactions` optionally filters `WHERE (category IS NULL OR category = '')` in addition to `user_id = ? AND bank_id = ?`. For users with many transactions, this OR-condition on `category` can't use a simple btree index efficiently in combination with the equality filters via a single index scan in all cases. A partial index covering uncategorized rows scoped to the existing `(user_id, bank_id)` access pattern resolves this.
**Recommended migration:** see `migrations/005_add_missing_indexes.sql`.

---

## 5. Schema drift: Go models vs SQL schema

| Model | Field | SQL column / type | Drift? |
|---|---|---|---|
| `models.User` | embeds `gorm.Model` (adds `ID uint` PK, `CreatedAt`, `UpdatedAt`, `DeletedAt`) **and** redeclares `ID string` | `users.id VARCHAR(255)` | **Finding F7** — `gorm.Model`'s own `ID uint` field is shadowed by the explicit `ID string` field, which works in Go/GORM (explicit field wins) but is fragile/confusing; the embedded `gorm.Model.CreatedAt/UpdatedAt/DeletedAt` are *not* explicitly redeclared with `json`/`db` tags matching the snake_case SQL columns, relying on GORM's default naming convention to match `created_at`/`updated_at`/`deleted_at`. This works but is undocumented/implicit. No immediate functional drift found, but it's a fragile pattern. |
| `models.Transaction` | `BankId string` `gorm:"type:uuid"` | `transactions.bank_id VARCHAR(255)` | **Finding F4 (High)** — type mismatch. Go/GORM annotates the column as `uuid` but the actual column is `VARCHAR(255)`, and seed data in `bank_accounts` uses non-UUID string IDs (`'scotiabank'`, `'nubank'`, `'wise'`, etc.). All transaction queries (`transaction.go` in postgresRepository) explicitly cast the parameter with `?::uuid` in raw SQL fragments (e.g. `"user_id = ? AND bank_id = ?::uuid"`), and `api/pkg/handlers/transaction/list.go` validates `BankId` as `binding:"required,uuid"` at the HTTP layer. **This means the current codebase requires bank_id values to be UUIDs, but the seed data inserted by `002_create_bank_accounts.sql` are human-readable strings, not UUIDs — these queries would fail (`invalid input syntax for type uuid`) against the seeded data as it exists today.** This is the most significant drift finding in this audit. |
| `models.Transaction` | `Fingerprint string` `gorm:"uniqueIndex:idx_transaction_fingerprint"` | **No `fingerprint` column in `run_migrations.sql`** | **Finding F8 (High)** — `CreateUniqueTransactionInBatches` does an `ON CONFLICT (fingerprint) DO NOTHING`, and the model declares a unique index via GORM tag, but no migration file creates the `fingerprint` column or its unique index on `transactions`. This column must currently be relying on `AutoMigrate` (`MigrateTransaction` in `postgresRepository/transaction.go` calls `p.db.AutoMigrate(model)`) to silently add it outside of the tracked migration files. This is schema drift between what's reviewable in `migrations/` and what's actually applied in any environment where `AutoMigrate` has run. |
| `models.BankAccount` | matches `bank_accounts` columns 1:1 (id, name, display_name, country_code, currency, created_at, updated_at, is_active) | OK | No drift |
| `models.UserBankAccount` | matches `user_bank_accounts` columns 1:1 | OK | No drift |
| `models.User.FinancialAppCredentials` / `BankCredentials` | `gorm:"type:jsonb"` | `users.financial_app_credentials JSONB`, `users.bank_credentials JSONB` | OK |

### Finding F4 — `bank_id` type mismatch: VARCHAR(255) in SQL vs UUID-cast in Go/SQL queries and seed data is non-UUID
**Severity:** High
**Description:** See table above. `bank_accounts.id` (and therefore everywhere `bank_id` is stored/joined) is `VARCHAR(255)` and seeded with values like `'scotiabank'`, `'nubank'`, `'wise'`, `'firsttech'`. However:
- `models.Transaction.BankId` is tagged `gorm:"type:uuid"`.
- Every transaction repository query that filters by `bank_id` casts the bind parameter to `::uuid` (`transaction.go` lines 34, 46, 56).
- The `ListTransactions` HTTP handler validates the `bankId` URI param with Gin's `binding:"required,uuid"`.
- `transactions.bank_id` has **no foreign key** to `bank_accounts.id` at all (unlike `user_bank_accounts.bank_id`, which does have the FK).

This is a pre-existing, high-severity inconsistency: either (a) the intent is for `bank_id` to actually be a UUID-typed surrogate key and the seed data / `bank_accounts.id` PK choice is wrong, or (b) the intent is for `bank_id` to be the human-readable string id and the `::uuid` casts / GORM `uuid` tag / Gin `uuid` validator are wrong and would currently reject/fail on valid bank ids such as `"nubank"`. This audit does not attempt to silently "fix" this by guessing intent, since it changes application behavior, not just schema metadata — it is called out here as the top finding for a maintainer to resolve deliberately. No migration is added for this; recommend either renaming `bank_accounts.id` semantics or fixing the Go-side `uuid` assumptions, plus adding the missing `transactions.bank_id → bank_accounts.id` FK once the type question is settled.

### Finding F8 — `fingerprint` column / unique index missing from tracked migrations
**Severity:** High
**Description:** See table above. `MigrateTransaction` calls `gorm.AutoMigrate`, which means schema changes can land in any deployed database without ever being captured as a reviewable, repeatable SQL migration file. This is a process/drift risk independent of whether the column exists today in any given environment — the migrations directory should be the source of truth.
**Recommended migration:** see `migrations/006_add_transaction_fingerprint.sql` (added in this PR) — defines the `fingerprint` column and unique index explicitly so the tracked migrations match what `AutoMigrate` has presumably already created live. **This migration is additive/idempotent (`IF NOT EXISTS`) and safe to run whether or not `AutoMigrate` already created the column.**

---

## Summary of new migration files added by this audit

| File | Purpose | Severity addressed |
|---|---|---|
| `migrations/004_user_bank_accounts_rls_hardening.sql` | Adds `FORCE ROW LEVEL SECURITY` and service-role bypass policies to `user_bank_accounts`, and a service-role write policy to `bank_accounts` | F2 (Medium), F3 (Low) |
| `migrations/005_add_missing_indexes.sql` | Adds composite index on `user_bank_accounts(user_id, is_active)`, `(bank_id, user_id, is_active)`, and a partial index on `transactions` for the uncategorized-filter query path | F5 (Low), F6 (Low) |
| `migrations/006_add_transaction_fingerprint.sql` | Explicitly defines `transactions.fingerprint` column + unique index that the Go model/repository already depend on via `AutoMigrate`, bringing it under migration tracking | F8 (High, process) |

**Not auto-fixed (requires a product/engineering decision, not added as a migration):**
- F1 / F1b — migration numbering collision and `run_migrations.sql` not including `002_create_bank_accounts.sql`. Recommend follow-up to adopt Supabase CLI's standard migrations directory convention.
- F4 — `bank_id` VARCHAR vs UUID mismatch between SQL schema, seed data, and Go/Gin UUID assumptions. This is a behavioral question (what should `bank_id` actually be?), not a safe mechanical fix.

All new migration files are additive only (`CREATE INDEX IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`, `DROP POLICY IF EXISTS` + recreate), safe to apply to an existing database, and have **not** been executed against any live database as part of this audit.
