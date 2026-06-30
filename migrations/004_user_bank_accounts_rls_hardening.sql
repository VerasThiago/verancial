-- Migration 004: Harden RLS on user_bank_accounts and bank_accounts
-- Adds FORCE ROW LEVEL SECURITY + service-role bypass policy to user_bank_accounts
-- (bringing it in line with the pattern already used on users/transactions),
-- and a service-role write policy to bank_accounts.
--
-- Safe to run multiple times: uses IF NOT EXISTS / DROP POLICY IF EXISTS guards.
-- NOTE: Not executed against any live database as part of this audit.

-- =====================================
-- USER_BANK_ACCOUNTS: force RLS + service-role bypass
-- =====================================

ALTER TABLE user_bank_accounts FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "service_role_user_bank_accounts" ON user_bank_accounts;
CREATE POLICY "service_role_user_bank_accounts" ON user_bank_accounts
  FOR ALL
  USING (current_setting('role') = 'service_role')
  WITH CHECK (current_setting('role') = 'service_role');

-- =====================================
-- BANK_ACCOUNTS: service-role write policy
-- (existing "bank_accounts_select_active" policy is left untouched)
-- =====================================

DROP POLICY IF EXISTS "service_role_bank_accounts" ON bank_accounts;
CREATE POLICY "service_role_bank_accounts" ON bank_accounts
  FOR ALL
  USING (current_setting('role') = 'service_role')
  WITH CHECK (current_setting('role') = 'service_role');
