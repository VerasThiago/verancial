-- Migration 006: Explicitly track the `fingerprint` column on transactions.
--
-- shared/models/transaction.go declares:
--   Fingerprint string `json:"fingerprint" gorm:"uniqueIndex:idx_transaction_fingerprint"`
-- and shared/repository/postgresRepository/transaction.go's
-- CreateUniqueTransactionInBatches relies on `ON CONFLICT (fingerprint) DO NOTHING`,
-- and MigrateTransaction calls gorm.AutoMigrate(model) which can silently add this
-- column outside of tracked migrations. This migration brings it under explicit,
-- reviewable SQL so the migrations/ directory matches what the application
-- actually depends on, regardless of whether AutoMigrate has already created it
-- in a given environment.
--
-- Idempotent: safe to run whether or not the column/index already exist.
-- NOTE: Not executed against any live database as part of this audit.

ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transaction_fingerprint
  ON transactions(fingerprint);
