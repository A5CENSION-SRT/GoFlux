package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		fmt.Println("Error pinging database:", err)
		return
	}
	fmt.Println("Database connected successfully")

	customerIDs, err := seedCustomers(ctx, pool)
	if err != nil {
		fmt.Println("Error seeding customers:", err)
		return
	}
	fmt.Printf("Seeded %d customers\n", len(customerIDs))

	categoryIDs, err := seedMerchantCategories(ctx, pool)
	if err != nil {
		fmt.Println("Error seeding merchant categories:", err)
		return
	}
	fmt.Printf("Seeded %d merchant categories\n", len(categoryIDs))

	merchantIDs, err := seedMerchants(ctx, pool, categoryIDs)
	if err != nil {
		fmt.Println("Error seeding merchants:", err)
		return
	}
	fmt.Printf("Seeded %d merchants\n", len(merchantIDs))

	accountIDs, err := seedAccounts(ctx, pool, customerIDs)
	if err != nil {
		fmt.Println("Error seeding accounts:", err)
		return
	}
	fmt.Printf("Seeded %d accounts\n", len(accountIDs))

	err = seedAccountBalances(ctx, pool, accountIDs)
	if err != nil {
		fmt.Println("Error seeding account balances:", err)
		return
	}
	fmt.Printf("Seeded %d account balances\n", len(accountIDs))

	transactionIDs, err := seedTransactions(ctx, pool, accountIDs, merchantIDs, categoryIDs)
	if err != nil {
		fmt.Println("Error seeding transactions:", err)
		return
	}
	fmt.Printf("Seeded %d transactions\n", len(transactionIDs))

	err = seedTransactionMetadata(ctx, pool, transactionIDs)
	if err != nil {
		fmt.Println("Error seeding transaction metadata:", err)
		return
	}
	fmt.Println("Seeded transaction metadata")

	transferIDs, err := seedTransfers(ctx, pool, accountIDs)
	if err != nil {
		fmt.Println("Error seeding transfers:", err)
		return
	}
	fmt.Printf("Seeded %d transfers\n", len(transferIDs))

	err = seedTransferHops(ctx, pool, transferIDs)
	if err != nil {
		fmt.Println("Error seeding transfer hops:", err)
		return
	}
	fmt.Println("Seeded transfer hops")

	batchIDs, err := seedSettlementBatches(ctx, pool)
	if err != nil {
		fmt.Println("Error seeding settlement batches:", err)
		return
	}
	fmt.Printf("Seeded %d settlement batches\n", len(batchIDs))

	err = seedSettlementBatchItems(ctx, pool, batchIDs, transferIDs)
	if err != nil {
		fmt.Println("Error seeding settlement batch items:", err)
		return
	}
	fmt.Println("Seeded settlement batch items")

	err = seedBudgets(ctx, pool, accountIDs, categoryIDs)
	if err != nil {
		fmt.Println("Error seeding budgets:", err)
		return
	}
	fmt.Println("Seeded budgets")

	fmt.Println("Database seeded successfully")
}

//  Customers 

func seedCustomers(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	batch := &pgx.Batch{}
	total := 10000

	for i := 0; i < total; i++ {
		batch.Queue(`
			INSERT INTO customers (full_name, email, phone, country_code, kyc_status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`,
			gofakeit.Name(),
			gofakeit.Email(),
			gofakeit.Phone(),
			gofakeit.CountryAbr(),
			"pending",
		)
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	var ids []string
	for i := 0; i < total; i++ {
		var id string
		err := results.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Merchant Categories 

func seedMerchantCategories(ctx context.Context, pool *pgxpool.Pool) (map[string]string, error) {
	topLevel := []struct {
		name  string
		icon  string
		color string
	}{
		{"Food & Dining", "fork-knife", "#FF5733"},
		{"Travel", "plane", "#2563EB"},
		{"Shopping", "bag", "#10B981"},
		{"Entertainment", "film", "#8B5CF6"},
		{"Healthcare", "heart", "#EF4444"},
		{"Utilities", "bolt", "#F59E0B"},
		{"Education", "book", "#3B82F6"},
		{"Transport", "car", "#6B7280"},
	}

	subCategories := map[string][]string{
		"Food & Dining":  {"Restaurants", "Groceries", "Fast Food", "Cafes"},
		"Travel":         {"Flights", "Hotels", "Car Rental"},
		"Shopping":       {"Clothing", "Electronics", "Home & Garden"},
		"Entertainment":  {"Movies", "Music", "Gaming"},
		"Healthcare":     {"Pharmacy", "Hospital", "Gym"},
		"Utilities":      {"Electricity", "Water", "Internet"},
		"Education":      {"Courses", "Books", "Stationery"},
		"Transport":      {"Fuel", "Parking", "Public Transit"},
	}

	categoryIDs := make(map[string]string)

	// insert top level first — subcategories need these IDs as parent_id
	for _, cat := range topLevel {
		var id string
		err := pool.QueryRow(ctx, `
			INSERT INTO merchant_categories (name, icon, color)
			VALUES ($1, $2, $3)
			RETURNING id
		`, cat.name, cat.icon, cat.color).Scan(&id)
		if err != nil {
			return nil, err
		}
		categoryIDs[cat.name] = id
	}

	// insert subcategories using parent IDs 
	for parentName, subs := range subCategories {
		parentID := categoryIDs[parentName]
		for _, subName := range subs {
			var id string
			err := pool.QueryRow(ctx, `
				INSERT INTO merchant_categories (name, parent_id, icon, color)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`, subName, parentID, "tag", "#9CA3AF").Scan(&id)
			if err != nil {
				return nil, err
			}
			categoryIDs[subName] = id
		}
	}

	return categoryIDs, nil
}

// Merchants 

func seedMerchants(ctx context.Context, pool *pgxpool.Pool, categoryIDs map[string]string) ([]string, error) {
	subCategoryNames := []string{
		"Restaurants", "Groceries", "Fast Food", "Cafes",
		"Flights", "Hotels", "Car Rental",
		"Clothing", "Electronics", "Home & Garden",
		"Movies", "Music", "Gaming",
		"Pharmacy", "Hospital", "Gym",
		"Electricity", "Water", "Internet",
		"Courses", "Books", "Stationery",
		"Fuel", "Parking", "Public Transit",
	}

	batch := &pgx.Batch{}
	total := 500

	for i := 0; i < total; i++ {
		subCatName := subCategoryNames[i%len(subCategoryNames)]
		categoryID := categoryIDs[subCatName]

		batch.Queue(`
			INSERT INTO merchants (name, category_id, country_code, mcc_code)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`,
			gofakeit.Company(),
			categoryID,
			gofakeit.CountryAbr(),
			fmt.Sprintf("%04d", gofakeit.Number(1000, 9999)),
		)
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	var ids []string
	for i := 0; i < total; i++ {
		var id string
		err := results.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Accounts 

func seedAccounts(ctx context.Context, pool *pgxpool.Pool, customerIDs []string) ([]string, error) {
	accountTypes := []string{"checking", "savings", "wallet"}
	batch := &pgx.Batch{}
	total := len(customerIDs) * 3

	for _, customerID := range customerIDs {
		for _, accountType := range accountTypes {
			batch.Queue(`
				INSERT INTO accounts (customer_id, account_type, currency, status)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`,
				customerID,
				accountType,
				"INR",
				"active",
			)
		}
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	var ids []string
	for i := 0; i < total; i++ {
		var id string
		err := results.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Account Balances 

func seedAccountBalances(ctx context.Context, pool *pgxpool.Pool, accountIDs []string) error {
	batch := &pgx.Batch{}

	for _, accountID := range accountIDs {
		batch.Queue(`
			INSERT INTO account_balances (account_id, available_balance, pending_balance)
			VALUES ($1, $2, $3)
		`,
			accountID,
			gofakeit.Float64Range(1000, 500000),
			gofakeit.Float64Range(0, 10000),
		)
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	for range accountIDs {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// Transactions 

func seedTransactions(ctx context.Context, pool *pgxpool.Pool, accountIDs []string, merchantIDs []string, categoryIDs map[string]string) ([]string, error) {
	directions := []string{"debit", "credit"}
	statuses := []string{"pending", "completed", "completed", "completed", "failed"}

	var catIDs []string
	for _, id := range categoryIDs {
		catIDs = append(catIDs, id)
	}

	total := 500000
	batchSize := 1000
	var allIDs []string

	for start := 0; start < total; start += batchSize {
		batch := &pgx.Batch{}
		end := start + batchSize
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			accountID := accountIDs[i%len(accountIDs)]
			merchantID := merchantIDs[i%len(merchantIDs)]
			categoryID := catIDs[i%len(catIDs)]
			direction := directions[i%len(directions)]
			status := statuses[i%len(statuses)]

			batch.Queue(`
				INSERT INTO transactions (account_id, merchant_id, category_id, amount, currency, direction, status, description, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id
			`,
				accountID,
				merchantID,
				categoryID,
				gofakeit.Float64Range(10, 50000),
				"INR",
				direction,
				status,
				gofakeit.Sentence(5),
				gofakeit.DateRange(
					mustParseTime("2025-01-01"),
					mustParseTime("2026-12-31"),
				),
			)
		}

		results := pool.SendBatch(ctx, batch)
		for i := start; i < end; i++ {
			var id string
			err := results.QueryRow().Scan(&id)
			if err != nil {
				results.Close()
				return nil, err
			}
			allIDs = append(allIDs, id)
		}
		results.Close()
		fmt.Printf("inserted transactions %d/%d\n", end, total)
	}
	return allIDs, nil
}

// Transaction Metadata 

func seedTransactionMetadata(ctx context.Context, pool *pgxpool.Pool, transactionIDs []string) error {
	metadataKeys := []struct {
		key   string
		value func() string
	}{
		{"card_last_four", func() string { return fmt.Sprintf("%04d", gofakeit.Number(1000, 9999)) }},
		{"terminal_id", func() string { return fmt.Sprintf("TERM-%04d", gofakeit.Number(1000, 9999)) }},
		{"authorization_code", func() string { return fmt.Sprintf("AUTH-%s", gofakeit.LetterN(6)) }},
		{"merchant_ref", func() string { return gofakeit.UUID() }},
		{"gateway", func() string {
			gateways := []string{"razorpay", "stripe", "paytm", "phonepe"}
			return gateways[gofakeit.Number(0, len(gateways)-1)]
		}},
	}

	batchSize := 1000
	total := len(transactionIDs)

	for start := 0; start < total; start += batchSize {
		batch := &pgx.Batch{}
		end := start + batchSize
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			meta := metadataKeys[i%len(metadataKeys)]
			batch.Queue(`
				INSERT INTO transaction_metadata (transaction_id, key, value)
				VALUES ($1, $2, $3)
			`,
				transactionIDs[i],
				meta.key,
				meta.value(),
			)
		}

		results := pool.SendBatch(ctx, batch)
		for range transactionIDs[start:end] {
			_, err := results.Exec()
			if err != nil {
				results.Close()
				return err
			}
		}
		results.Close()
		fmt.Printf("inserted transaction metadata %d/%d\n", end, total)
	}
	return nil
}

// Transfers 

func seedTransfers(ctx context.Context, pool *pgxpool.Pool, accountIDs []string) ([]string, error) {
	statuses := []string{"initiated", "routing", "settled", "settled", "settled", "failed"}
	total := 50000
	batchSize := 1000
	var ids []string

	for start := 0; start < total; start += batchSize {
		batch := &pgx.Batch{}
		end := start + batchSize
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			originID := accountIDs[i%len(accountIDs)]
			destID := accountIDs[(i+1)%len(accountIDs)]
			status := statuses[i%len(statuses)]

			batch.Queue(`
				INSERT INTO transfers (origin_account_id, destination_account_id, amount, currency, status, exchange_rate, initiated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`,
				originID,
				destID,
				gofakeit.Float64Range(100, 100000),
				"INR",
				status,
				gofakeit.Float64Range(0.8, 1.2),
				gofakeit.DateRange(
					mustParseTime("2025-01-01"),
					mustParseTime("2026-12-31"),
				),
			)
		}

		results := pool.SendBatch(ctx, batch)
		for i := start; i < end; i++ {
			var id string
			err := results.QueryRow().Scan(&id)
			if err != nil {
				results.Close()
				return nil, err
			}
			ids = append(ids, id)
		}
		results.Close()
		fmt.Printf("inserted transfers %d/%d\n", end, total)
	}
	return ids, nil
}

// Transfer Hops

func seedTransferHops(ctx context.Context, pool *pgxpool.Pool, transferIDs []string) error {
	hopStatuses := []string{"pending", "processing", "completed", "completed", "failed"}
	nodes := []string{"HDFC-IN", "SBI-IN", "AXIS-IN", "DEUTSCHE-DE", "HSBC-UK", "CITI-US", "DBS-SG"}

	batchSize := 1000
	var allHops []struct {
		transferID string
		seq        int
		from       string
		to         string
		status     string
	}

	// generate 1-3 hops per transfer
	for i, transferID := range transferIDs {
		hopCount := (i % 3) + 1
		for seq := 1; seq <= hopCount; seq++ {
			allHops = append(allHops, struct {
				transferID string
				seq        int
				from       string
				to         string
				status     string
			}{
				transferID: transferID,
				seq:        seq,
				from:       nodes[i%len(nodes)],
				to:         nodes[(i+seq)%len(nodes)],
				status:     hopStatuses[i%len(hopStatuses)],
			})
		}
	}

	total := len(allHops)
	for start := 0; start < total; start += batchSize {
		batch := &pgx.Batch{}
		end := start + batchSize
		if end > total {
			end = total
		}

		for _, hop := range allHops[start:end] {
			batch.Queue(`
				INSERT INTO transfer_hops (transfer_id, sequence_number, from_node, to_node, hop_status, amount, fee)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`,
				hop.transferID,
				hop.seq,
				hop.from,
				hop.to,
				hop.status,
				gofakeit.Float64Range(100, 100000),
				gofakeit.Float64Range(0, 500),
			)
		}

		results := pool.SendBatch(ctx, batch)
		for range allHops[start:end] {
			_, err := results.Exec()
			if err != nil {
				results.Close()
				return err
			}
		}
		results.Close()
		fmt.Printf("inserted transfer hops %d/%d\n", end, total)
	}
	return nil
}

// Settlement Batches 

func seedSettlementBatches(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	batch := &pgx.Batch{}

	// one batch per day for all of 2025
	startDate := mustParseTime("2025-01-01")
	endDate := mustParseTime("2025-12-31")
	total := 0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		batch.Queue(`
			INSERT INTO settlement_batches (batch_date, status, total_amount, transfer_count)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`,
			d.Format("2006-01-02"),
			"settled",
			gofakeit.Float64Range(100000, 10000000),
			gofakeit.Number(10, 500),
		)
		total++
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	var ids []string
	for i := 0; i < total; i++ {
		var id string
		err := results.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Settlement Batch Items

func seedSettlementBatchItems(ctx context.Context, pool *pgxpool.Pool, batchIDs []string, transferIDs []string) error {
	// use settled transfers only — take first 30000
	settledTransfers := transferIDs
	if len(settledTransfers) > 30000 {
		settledTransfers = transferIDs[:30000]
	}

	batchSize := 1000
	total := len(settledTransfers)

	for start := 0; start < total; start += batchSize {
		batch := &pgx.Batch{}
		end := start + batchSize
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			batchID := batchIDs[i%len(batchIDs)]
			transferID := settledTransfers[i]

			batch.Queue(`
				INSERT INTO settlement_batch_items (batch_id, transfer_id, amount, status)
				VALUES ($1, $2, $3, $4)
			`,
				batchID,
				transferID,
				gofakeit.Float64Range(100, 100000),
				"settled",
			)
		}

		results := pool.SendBatch(ctx, batch)
		for range settledTransfers[start:end] {
			_, err := results.Exec()
			if err != nil {
				results.Close()
				return err
			}
		}
		results.Close()
		fmt.Printf("inserted settlement batch items %d/%d\n", end, total)
	}
	return nil
}

// Budgets 
func seedBudgets(ctx context.Context, pool *pgxpool.Pool, accountIDs []string, categoryIDs map[string]string) error {
	var catIDs []string
	for _, id := range categoryIDs {
		catIDs = append(catIDs, id)
	}

	batch := &pgx.Batch{}
	total := 10000

	for i := 0; i < total; i++ {
		accountID := accountIDs[i%len(accountIDs)]
		categoryID := catIDs[i%len(catIDs)]

		batch.Queue(`
			INSERT INTO budgets (account_id, category_id, amount, period, start_date, end_date, alert_threshold)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			accountID,
			categoryID,
			gofakeit.Float64Range(1000, 50000),
			"monthly",
			"2026-01-01",
			"2026-12-31",
			80.00,
		)
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < total; i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}