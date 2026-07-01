-- Migration 002: Create Bank Accounts and User Bank Accounts tables
-- Run this after the main migration script

-- Create Bank Accounts table (supported banks)
-- id is UUID (not a slug) to match shared/constants/bank_name.go and the
-- `bank_id = ?::uuid` casts used throughout shared/repository/postgresRepository.
CREATE TABLE IF NOT EXISTS bank_accounts (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    country_code VARCHAR(3),
    currency VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Create User Bank Accounts relation table
CREATE TABLE IF NOT EXISTS user_bank_accounts (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    user_id VARCHAR(255) NOT NULL,
    bank_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    last_sync_date TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (bank_id) REFERENCES bank_accounts(id) ON DELETE CASCADE,
    UNIQUE(user_id, bank_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_bank_accounts_active ON bank_accounts(is_active);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_user_id ON user_bank_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_bank_id ON user_bank_accounts(bank_id);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_active ON user_bank_accounts(is_active);

-- Insert supported bank accounts (UUIDs matching shared/constants/bank_name.go)
INSERT INTO bank_accounts (id, name, display_name, country_code, currency) VALUES
('92f1aff9-03ae-4e6e-bde2-7799e849d181', 'Scotia Bank', 'Scotia Bank', 'CAN', 'CAD'),
('8462037f-7615-406a-b8fc-214105becd33', 'Scotia Bank Credit Card', 'Scotia Bank Credit Card', 'CAN', 'CAD'),
('b5604a4f-1389-45ae-af18-915acc268fed', 'Nubank', 'Nubank', 'BRA', 'BRL'),
('5ae0c7ff-20c2-4bb8-af55-35347df9a9fd', 'Wise', 'Wise', 'INT', 'CAD'),
('91c1ae86-05e7-4b57-b288-cd6fe5a61ccb', 'First Tech Federal Credit Union', 'First Tech Federal Credit Union', 'USA', 'USD'),
('a317dc85-81bb-40ad-a12f-d8b546c1230b', 'First Tech Credit Card', 'First Tech Credit Card', 'USA', 'USD')
ON CONFLICT (id) DO NOTHING;

-- Enable Row Level Security
ALTER TABLE bank_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_bank_accounts ENABLE ROW LEVEL SECURITY;

-- =====================================
-- BANK ACCOUNTS TABLE POLICIES
-- =====================================

-- Policy: Anyone can view active bank accounts (for registration/connection)
CREATE POLICY "bank_accounts_select_active" ON bank_accounts
  FOR SELECT
  USING (is_active = true);

-- =====================================
-- USER BANK ACCOUNTS TABLE POLICIES
-- =====================================

-- Policy: Users can view their own bank account connections
CREATE POLICY "user_bank_accounts_select_own" ON user_bank_accounts
  FOR SELECT
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can connect their own bank accounts
CREATE POLICY "user_bank_accounts_insert_own" ON user_bank_accounts
  FOR INSERT
  WITH CHECK (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can update their own bank account connections
CREATE POLICY "user_bank_accounts_update_own" ON user_bank_accounts
  FOR UPDATE
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can delete their own bank account connections
CREATE POLICY "user_bank_accounts_delete_own" ON user_bank_accounts
  FOR DELETE
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub'); 