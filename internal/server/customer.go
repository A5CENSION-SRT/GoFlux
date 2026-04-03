package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "github.com/A5CENSION-SRT/goflux/internal/gen/goflux/v1"
)

// the concrete service.CustomerService 
type CustomerServiceInterface interface {
	CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.Customer, error)
	GetCustomer(ctx context.Context, customerID string) (*pb.Customer, error)
	UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.Customer, error)
	UpdateKYCStatus(ctx context.Context, req *pb.UpdateKYCStatusRequest) (*pb.Customer, error)
	DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) error
	ListCustomers(ctx context.Context, req *pb.ListCustomersRequest) ([]*pb.Customer, int32, error)
}

type CustomerHandler struct {
	pb.UnimplementedCustomerServiceServer
	service CustomerServiceInterface
}

func NewCustomerHandler(service CustomerServiceInterface) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error) {
	if req.FullName == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name and email are required")
	}

	customer, err := h.service.CreateCustomer(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create customer: %v", err)
	}

	return &pb.CreateCustomerResponse{Customer: customer}, nil
}

func (h *CustomerHandler) GetCustomer(ctx context.Context, req *pb.GetCustomerRequest) (*pb.GetCustomerResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	customer, err := h.service.GetCustomer(ctx, req.CustomerId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "customer not found: %v", err)
	}

	return &pb.GetCustomerResponse{Customer: customer}, nil
}

func (h *CustomerHandler) UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.UpdateCustomerResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	customer, err := h.service.UpdateCustomer(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update customer: %v", err)
	}

	return &pb.UpdateCustomerResponse{Customer: customer}, nil
}

func (h *CustomerHandler) UpdateKYCStatus(ctx context.Context, req *pb.UpdateKYCStatusRequest) (*pb.UpdateKYCStatusResponse, error) {
	if req.CustomerId == "" || req.KycStatus == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id and kyc_status are required")
	}

	customer, err := h.service.UpdateKYCStatus(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update kyc status: %v", err)
	}

	return &pb.UpdateKYCStatusResponse{Customer: customer}, nil
}

func (h *CustomerHandler) DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) (*pb.DeleteCustomerResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	err := h.service.DeleteCustomer(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete customer: %v", err)
	}

	return &pb.DeleteCustomerResponse{Success: true, Message: "customer successfully closed"}, nil
}

func (h *CustomerHandler) ListCustomers(ctx context.Context, req *pb.ListCustomersRequest) (*pb.ListCustomersResponse, error) {
	customers, total, err := h.service.ListCustomers(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list customers: %v", err)
	}

	return &pb.ListCustomersResponse{
		Customers: customers,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}