-- Migration 007: Fix transactions.bank_id column type (VARCHAR -> UUID)
--
-- transactions.bank_id was left as VARCHAR(255) in run_migrations.sql, but
-- every read in shared/repository/postgresRepository/transaction.go casts
-- the query parameter as `bank_id = ?::uuid`, e.g.:
--   GetLastTransactionFromUserBank, GetAllTransactionsFromUserBankAfterDate,
--   GetTransactions, GetTransactionCountFromUserBank
-- Postgres has no `character varying = uuid` operator, so every one of
-- those queries fails outright with "operator does not exist: character
-- varying = uuid" (SQLSTATE 42883) against a real database. This was never
-- caught before because there was no test that ran these queries against
-- actual Postgres -- caught by the e2e workflow's first real run.
--
-- bank_accounts.id and user_bank_accounts.bank_id were already fixed to
-- UUID in migration 002 (see also PR #22); this brings transactions.bank_id
-- in line with the same convention the application code has always assumed.
--
-- USING bank_id::uuid requires every existing non-null value to already be
-- a valid UUID string, which matches what the app has always written here
-- (bank IDs from shared/constants/bank_name.go).

ALTER TABLE transactions
  ALTER COLUMN bank_id TYPE UUID USING bank_id::uuid;
