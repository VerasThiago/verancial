#!/bin/sh
# Applies the repo's real migrations/*.sql files, in dependency order, against
# the throwaway e2e Postgres instance. Runs as the postgres superuser, which
# bypasses RLS (including FORCE ROW LEVEL SECURITY) unconditionally, so no
# service_role setup is needed here.
set -eu

MIGRATIONS_DIR=/migrations

for f in \
  run_migrations.sql \
  002_create_bank_accounts.sql \
  004_user_bank_accounts_rls_hardening.sql \
  005_add_missing_indexes.sql \
  006_add_transaction_fingerprint.sql \
  007_fix_transactions_bank_id_type.sql
do
  echo "Applying $f..."
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$MIGRATIONS_DIR/$f"
done
