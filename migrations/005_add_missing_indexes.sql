-- Migration 005: Add missing indexes identified by cross-referencing
-- shared/repository/postgresRepository queries against existing indexes.
--
-- Safe to run multiple times: uses IF NOT EXISTS guards.
-- NOTE: Not executed against any live database as part of this audit.

-- Supports GetUserBankAccounts(userId) -> WHERE user_id = ? AND is_active = ?
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_user_active
  ON user_bank_accounts(user_id, is_active);

-- Supports GetBankAccountById(bankId, userId) -> WHERE bank_id = ? AND user_id = ? AND is_active = ?
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_bank_user_active
  ON user_bank_accounts(bank_id, user_id, is_active);

-- Supports GetTransactions(...) optional filter: WHERE user_id = ? AND bank_id = ?
-- AND (category IS NULL OR category = '')
-- Partial index covering only the "uncategorized" rows, scoped to the
-- existing (user_id, bank_id) access pattern.
CREATE INDEX IF NOT EXISTS idx_transactions_uncategorized
  ON transactions(user_id, bank_id, date DESC)
  WHERE (category IS NULL OR category = '');
