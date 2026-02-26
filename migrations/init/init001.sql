CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE account_type AS ENUM ('checking', 'savings', 'wallet');
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'suspended', 'closed');
CREATE TYPE transaction_direction AS ENUM ('debit', 'credit');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'reversed');
CREATE TYPE transfer_status AS ENUM ('initiated', 'routing', 'settled', 'failed');
CREATE TYPE hop_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE kyc_status AS ENUM ('pending', 'verified', 'rejected');
CREATE TYPE budget_period AS ENUM ('weekly', 'monthly', 'yearly');

-- Customers
CREATE TABLE customers (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name    VARCHAR(255) NOT NULL,
    email        VARCHAR(255) NOT NULL UNIQUE,
    phone        VARCHAR(20) NOT NULL UNIQUE,
    country_code CHAR(2) NOT NULL,
    kyc_status   kyc_status NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW() 
);

-- Accounts
CREATE TABLE accounts(
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    account_type account_type NOT NULL,
    currency     CHAR(3) NOT NULL DEFAULT 'INR',
    status       account_status NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW() 
);

CREATE INDEX idx_accounts_customer_status ON accounts(customer_id, status);

-- Account Balances
CREATE TABLE account_balances (
    account_id        UUID PRIMARY KEY REFERENCES accounts(id) ON DELETE RESTRICT,
    available_balance NUMERIC(19, 4) NOT NULL DEFAULT 0.0000,
    pending_balance   NUMERIC(19, 4) NOT NULL DEFAULT 0.0000,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--Merchants categories
CREATE TABLE merchant_categories (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    parent_id  UUID REFERENCES merchant_categories(id) ON DELETE RESTRICT,
    icon       VARCHAR(50),
    color      CHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchant_categories_parent_id ON merchant_categories(parent_id);

-- Merchants
CREATE TABLE merchants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    category_id     UUID REFERENCES merchant_categories(id) ON DELETE RESTRICT,
    country_code    CHAR(2) NOT NULL,
    mcc_code        CHAR(4),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchants_category_id ON merchants(category_id);
CREATE INDEX idx_merchants_mcc_code ON merchants(mcc_code);

--Transactions
CREATE TABLE transactions (
    id              UUID NOT NULL DEFAULT uuid_generate_v4(),
    account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    merchant_id     UUID REFERENCES merchants(id) ON DELETE RESTRICT,
    category_id     UUID REFERENCES merchant_categories(id) ON DELETE RESTRICT,
    amount          NUMERIC(19, 4) NOT NULL,
    currency        CHAR(3) NOT NULL DEFAULT 'INR',
    direction       transaction_direction NOT NULL,
    status          transaction_status NOT NULL DEFAULT 'pending',
    reference       VARCHAR(100) UNIQUE,
    description     VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at      TIMESTAMPTZ,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for transactions table
CREATE TABLE transactions_2025 PARTITION OF transactions
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE transactions_2026 PARTITION OF transactions
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

CREATE INDEX idx_transactions_account_created ON transactions(account_id, created_at DESC);
CREATE INDEX idx_transactions_merchant_id ON transactions(merchant_id);
CREATE INDEX idx_transactions_pending ON transactions(status, created_at) WHERE status = 'pending';
CREATE INDEX idx_transactions_category_id ON transactions(category_id);

-- Transaction Metadata
CREATE TABLE transaction_metadata (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id  UUID NOT NULL,
    key             VARCHAR(100) NOT NULL,
    value           TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(transaction_id, key)
);

CREATE INDEX idx_transaction_metadata_transaction_key ON transaction_metadata(transaction_id, key);

-- Transfers
CREATE TABLE transfers (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    origin_account_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    destination_account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount                  NUMERIC(19, 4) NOT NULL,
    currency                CHAR(3) NOT NULL DEFAULT 'INR',
    status                  transfer_status NOT NULL DEFAULT 'initiated',
    exchange_rate           NUMERIC(19, 6),
    initiated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at              TIMESTAMPTZ,
    description             VARCHAR(500)
);

CREATE INDEX idx_transfers_origin ON transfers(origin_account_id, initiated_at DESC);
CREATE INDEX idx_transfers_destination ON transfers(destination_account_id, initiated_at DESC);
CREATE INDEX idx_transfers_status ON transfers(status) WHERE status IN ('initiated', 'routing');

-- Transfer Hops
CREATE TABLE transfer_hops (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transfer_id     UUID NOT NULL REFERENCES transfers(id) ON DELETE RESTRICT,
    sequence_number INTEGER NOT NULL,
    from_node       VARCHAR(255) NOT NULL,
    to_node         VARCHAR(255) NOT NULL,
    hop_status      hop_status NOT NULL DEFAULT 'pending',
    amount          NUMERIC(19, 4) NOT NULL,
    fee             NUMERIC(19, 4) NOT NULL DEFAULT 0.0000,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(transfer_id, sequence_number)
);

CREATE INDEX idx_transfer_hops_transfer_sequence ON transfer_hops(transfer_id, sequence_number);
CREATE INDEX idx_transfer_hops_status ON transfer_hops(hop_status) WHERE hop_status = 'processing';

-- Settlement Batches
CREATE TABLE settlement_batches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_date      DATE NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'open',
    total_amount    NUMERIC(19, 4) NOT NULL DEFAULT 0.0000,
    transfer_count  INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at      TIMESTAMPTZ
);

CREATE INDEX idx_settlement_batches_status ON settlement_batches(status) WHERE status = 'open';
CREATE INDEX idx_settlement_batches_date ON settlement_batches(batch_date DESC);

-- Settlement Batch Items
CREATE TABLE settlement_batch_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id        UUID NOT NULL REFERENCES settlement_batches(id) ON DELETE RESTRICT,
    transfer_id     UUID NOT NULL REFERENCES transfers(id) ON DELETE RESTRICT,
    amount          NUMERIC(19, 4) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(batch_id, transfer_id)
);

CREATE INDEX idx_settlement_batch_items_batch_id ON settlement_batch_items(batch_id);
CREATE INDEX idx_settlement_batch_items_transfer_id ON settlement_batch_items(transfer_id);

-- Budgets
CREATE TABLE budgets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    category_id     UUID REFERENCES merchant_categories(id) ON DELETE RESTRICT,
    amount          NUMERIC(19, 4) NOT NULL,
    period          budget_period NOT NULL DEFAULT 'monthly',
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    alert_threshold NUMERIC(5, 2) NOT NULL DEFAULT 80.00,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_account_period ON budgets(account_id, start_date, end_date);
CREATE INDEX idx_budgets_category_id ON budgets(category_id);
-- Idempotency Keys
CREATE TABLE idempotency_keys (
    key             VARCHAR(255) PRIMARY KEY,
    request_hash    VARCHAR(64) NOT NULL,
    response_payload TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'processing',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);