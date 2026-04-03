package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Customer struct {
	ID          string
	FullName    string
	Email       string
	Phone       string
	CountryCode string
	KycStatus   string
	CreatedAt   string
	UpdatedAt   string
}

type CreateCustomerParams struct {
	FullName    string
	Email       string
	Phone       string
	CountryCode string
}

type ListCustomersParams struct {
	KycStatus   *string
	CountryCode *string
	Page        int32
	PageSize    int32
}

type CustomerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

func (r *CustomerRepository) CreateCustomer(ctx context.Context, params CreateCustomerParams) (*Customer, error) {
	var c Customer
	err := r.pool.QueryRow(ctx, `
		INSERT INTO customers (full_name, email, phone, country_code, kyc_status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, full_name, email, phone, country_code, kyc_status,
		          created_at::text, updated_at::text
	`, params.FullName, params.Email, params.Phone, params.CountryCode).
		Scan(&c.ID, &c.FullName, &c.Email, &c.Phone, &c.CountryCode,
			&c.KycStatus, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	var c Customer
	err := r.pool.QueryRow(ctx, `
		SELECT id, full_name, email, phone, country_code, kyc_status,
		       created_at::text, updated_at::text
		FROM customers
		WHERE id = $1
	`, customerID).
		Scan(&c.ID, &c.FullName, &c.Email, &c.Phone, &c.CountryCode,
			&c.KycStatus, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, customerID, fullName, email, phone string) (*Customer, error) {
	var c Customer
	err := r.pool.QueryRow(ctx, `
		UPDATE customers
		SET full_name = $1, email = $2, phone = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, full_name, email, phone, country_code, kyc_status,
		          created_at::text, updated_at::text
	`, fullName, email, phone, customerID).
		Scan(&c.ID, &c.FullName, &c.Email, &c.Phone, &c.CountryCode,
			&c.KycStatus, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) UpdateKYCStatus(ctx context.Context, customerID, kycStatus string) (*Customer, error) {
	var c Customer
	err := r.pool.QueryRow(ctx, `
		UPDATE customers
		SET kyc_status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, full_name, email, phone, country_code, kyc_status,
		          created_at::text, updated_at::text
	`, kycStatus, customerID).
		Scan(&c.ID, &c.FullName, &c.Email, &c.Phone, &c.CountryCode,
			&c.KycStatus, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update kyc status: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE customers SET kyc_status = 'rejected', updated_at = NOW()
		WHERE id = $1
	`, customerID)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) ListCustomers(ctx context.Context, params ListCustomersParams) ([]*Customer, int32, error) {
	// build query dynamically based on which filters were provided
	query := `SELECT id, full_name, email, phone, country_code, kyc_status,
	                 created_at::text, updated_at::text
	          FROM customers WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM customers WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if params.KycStatus != nil {
		clause := fmt.Sprintf(" AND kyc_status = $%d", argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.KycStatus)
		argCount++
	}

	if params.CountryCode != nil {
		clause := fmt.Sprintf(" AND country_code = $%d", argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.CountryCode)
		argCount++
	}

	// get total count before applying pagination
	var total int32
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers count: %w", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, params.PageSize, (params.Page-1)*int32(params.PageSize))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	var customers []*Customer
	for rows.Next() {
		var c Customer
		err := rows.Scan(&c.ID, &c.FullName, &c.Email, &c.Phone,
			&c.CountryCode, &c.KycStatus, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan customer: %w", err)
		}
		customers = append(customers, &c)
	}

	return customers, total, nil
}