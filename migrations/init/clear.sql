TRUNCATE TABLE 
    budgets,
    idempotency_keys,
    settlement_batch_items,
    settlement_batches,
    transfer_hops,
    transfers,
    transaction_metadata,
    transactions_2025,
    transactions_2026,
    account_balances,
    accounts,
    merchants,
    merchant_categories,
    customers
RESTART IDENTITY CASCADE;