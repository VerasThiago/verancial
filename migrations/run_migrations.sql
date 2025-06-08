-- Verancial Database Migration Script for Supabase
-- Run this script in your Supabase SQL editor or via CLI

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Migration 001: Create Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    financial_app_credentials JSONB,
    bank_credentials JSONB
);

-- Create indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Migration 002: Create Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    amount REAL NOT NULL,
    payee VARCHAR(255),
    description TEXT,
    category VARCHAR(255),
    currency VARCHAR(10),
    bank_id VARCHAR(255),
    metadata JSONB,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for transactions table
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_bank_id ON transactions(bank_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_bank ON transactions(user_id, bank_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_bank_date ON transactions(user_id, bank_id, date);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);

-- Enable Row Level Security (RLS) for better security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

-- =====================================
-- USER TABLE POLICIES
-- =====================================

-- Policy: Users can view their own profile
CREATE POLICY "users_select_own" ON users
  FOR SELECT
  USING (id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can update their own profile
CREATE POLICY "users_update_own" ON users
  FOR UPDATE
  USING (id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Allow user registration (insert new users)
CREATE POLICY "users_insert_registration" ON users
  FOR INSERT
  WITH CHECK (true);

-- Policy: Users can delete their own account
CREATE POLICY "users_delete_own" ON users
  FOR DELETE
  USING (id = current_setting('request.jwt.claims', true)::json->>'sub');

-- =====================================
-- TRANSACTIONS TABLE POLICIES
-- =====================================

-- Policy: Users can view their own transactions
CREATE POLICY "transactions_select_own" ON transactions
  FOR SELECT
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can insert their own transactions
CREATE POLICY "transactions_insert_own" ON transactions
  FOR INSERT
  WITH CHECK (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can update their own transactions
CREATE POLICY "transactions_update_own" ON transactions
  FOR UPDATE
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- Policy: Users can delete their own transactions
CREATE POLICY "transactions_delete_own" ON transactions
  FOR DELETE
  USING (user_id = current_setting('request.jwt.claims', true)::json->>'sub');

-- =====================================
-- SERVICE ROLE BYPASS (Important!)
-- =====================================

-- Allow service role to bypass RLS for server-side operations
ALTER TABLE users FORCE ROW LEVEL SECURITY;
ALTER TABLE transactions FORCE ROW LEVEL SECURITY;

-- Create policies for service role access
CREATE POLICY "service_role_users" ON users
  FOR ALL
  USING (current_setting('role') = 'service_role')
  WITH CHECK (current_setting('role') = 'service_role');

CREATE POLICY "service_role_transactions" ON transactions
  FOR ALL
  USING (current_setting('role') = 'service_role')
  WITH CHECK (current_setting('role') = 'service_role'); 