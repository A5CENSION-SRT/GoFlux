package service

import (
	"context"
	"fmt"
	"github.com/A5CENSION-SRT/goflux/internal/db"
	pb "github.com/A5CENSION-SRT/goflux/internal/gen/goflux/v1"
)

// the concrete db.CustomerRepository satisfies this interface
type CustomerRepository interface {
	CreateCustomer(ctx context.Context, params db.CreateCustomerParams) (*db.Customer, error)
	GetCustomer(ctx context.Context, customerID string) (*db.Customer, error)
	UpdateCustomer(ctx context.Context, customerID, fullName, email, phone string) (*db.Customer, error)
	UpdateKYCStatus(ctx context.Context, customerID, kycStatus string) (*db.Customer, error)
	DeleteCustomer(ctx context.Context, customerID string) error
	ListCustomers(ctx context.Context, params db.ListCustomersParams) ([]*db.Customer, int32, error)
}

type CustomerService struct {
	repo CustomerRepository
}

func NewCustomerService(repo CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.Customer, error) {
	if req.FullName == "" || req.Email == "" || req.Phone == "" || req.CountryCode == "" {
		return nil, fmt.Errorf("full_name, email, phone, country_code are all required")
	}

	customer, err := s.repo.CreateCustomer(ctx, db.CreateCustomerParams{
		FullName:    req.FullName,
		Email:       req.Email,
		Phone:       req.Phone,
		CountryCode: req.CountryCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}

	return toProtoCustomer(customer), nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, customerID string) (*pb.Customer, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	customer, err := s.repo.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}

	return toProtoCustomer(customer), nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.Customer, error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	customer, err := s.repo.UpdateCustomer(ctx, req.CustomerId, req.FullName, req.Email, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}

	return toProtoCustomer(customer), nil
}

func (s *CustomerService) UpdateKYCStatus(ctx context.Context, req *pb.UpdateKYCStatusRequest) (*pb.Customer, error) {
	if req.CustomerId == "" || req.KycStatus == "" {
		return nil, fmt.Errorf("customer_id and kyc_status are required")
	}

	customer, err := s.repo.UpdateKYCStatus(ctx, req.CustomerId, req.KycStatus)
	if err != nil {
		return nil, fmt.Errorf("update kyc status: %w", err)
	}

	return toProtoCustomer(customer), nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) error {
	if req.CustomerId == "" {
		return fmt.Errorf("customer_id is required")
	}

	return s.repo.DeleteCustomer(ctx, req.CustomerId)
}

func (s *CustomerService) ListCustomers(ctx context.Context, req *pb.ListCustomersRequest) ([]*pb.Customer, int32, error) {
	params := db.ListCustomersParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.KycStatus != nil {
		params.KycStatus = req.KycStatus
	}
	if req.CountryCode != nil {
		params.CountryCode = req.CountryCode
	}

	// default pagination if not provided
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	customers, total, err := s.repo.ListCustomers(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	var protoCustomers []*pb.Customer
	for _, c := range customers {
		protoCustomers = append(protoCustomers, toProtoCustomer(c))
	}

	return protoCustomers, total, nil
}

// toProtoCustomer converts a db.Customer to a pb.Customer (protobuf representation)
func toProtoCustomer(c *db.Customer) *pb.Customer {
	return &pb.Customer{
		Id:          c.ID,
		FullName:    c.FullName,
		Email:       c.Email,
		Phone:       c.Phone,
		CountryCode: c.CountryCode,
		KycStatus:   c.KycStatus,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}