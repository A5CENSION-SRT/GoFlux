CustomerService
CreateCustomer      — register a new customer
GetCustomer         — get customer profile by ID
UpdateCustomer      — update name, phone, email
UpdateKYCStatus     — update KYC verification status
DeleteCustomer      — soft delete a customer (set status to closed)
ListCustomers       — list all customers with filters

AccountService
CreateAccount        — create a new account for a customer
GetAccount           — get account details by ID
GetAccountBalance    — get available and pending balance
ListCustomerAccounts — list all accounts for a customer
UpdateAccountStatus  — activate, suspend, or close an account
CloseAccount         — permanently close an account

TransactionService
CreateTransaction        — record a new transaction
GetTransaction           — get a single transaction by ID
ListAccountTransactions  — paginated transaction history
ReverseTransaction       — reverse a completed transaction

TransferService
InitiateTransfer      — start a transfer between two accounts
GetTransferStatus     — get current status of a transfer
GetTransferPath       — full hop by hop routing path
CancelTransfer        — cancel an initiated transfer before it routes
RetryTransfer         — retry a failed transfer

BudgetService
CreateBudget    — create a budget
GetBudgetStatus — how much of a budget has been spent
ListBudgets     — list all budgets for an account
UpdateBudget    — change amount or alert threshold
DeleteBudget    — remove a budget

MerchantService
GetMerchant      — get merchant details by ID
ListMerchants    — list merchants with category filter
SearchMerchants  — search by name or MCC code
CreateMerchant   — add a new merchant
UpdateMerchant   — update merchant details

AnalyticsService
GetSpendingByCategory    — spending breakdown across category hierarchy
GetMonthlySpendingTrend  — month over month spending
GetTopMerchants          — top merchants by spend over a time period
GetCashflowSummary       — total debits vs credits over a period
GetBudgetVsActual        — compare budget against real spending per category

SettlementService
GetSettlementBatch    — get a batch by date
ListSettlementBatches — list batches with status filter
GetBatchItems         — list all transfers in a batch
ProcessBatch          — manually trigger batch processing

MerchantCategoryService
GetCategory      — get a category by ID
ListCategories   — list all top level categories
GetSubCategories — get children of a category
CreateCategory   — add a new category
UpdateCategory   — update name, icon, color
DeleteCategory   — remove a category