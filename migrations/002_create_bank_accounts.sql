-- Migration 002: Create Bank Accounts and User Bank Accounts tables
-- Run this after the main migration script

-- Create Bank Accounts table (supported banks)
CREATE TABLE IF NOT EXISTS bank_accounts (
    id VARCHAR(255) PRIMARY KEY,
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
    bank_id VARCHAR(255) NOT NULL,
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

-- Insert supported bank accounts (matching the existing constants)
INSERT INTO bank_accounts (id, name, display_name, country_code, currency) VALUES
('scotiabank', 'Scotia Bank', 'CAN', 'CAD'),
('scotiabank_cc', 'Scotia Bank Credit Card', 'CAN', 'CAD'),
('nubank', 'Nubank', 'BRA', 'BRL'),
('wise', 'Wise', 'INT', 'CAD'),
('firsttech', 'First Tech Federal Credit Union', 'USA', 'USD'),
('firsttech_cc', 'First Tech Credit Card', 'USA', 'USD')
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