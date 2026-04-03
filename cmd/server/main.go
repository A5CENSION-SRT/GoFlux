package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/A5CENSION-SRT/goflux/internal/config"
	"github.com/A5CENSION-SRT/goflux/internal/db"
	grpcserver "github.com/A5CENSION-SRT/goflux/internal/server"
	"github.com/A5CENSION-SRT/goflux/internal/service"
	pb "github.com/A5CENSION-SRT/goflux/internal/gen/goflux/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"context"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background() // connect to database 
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// verify database is reachable 
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	fmt.Println("connected to database successfully")

	//layers — repository → service → handler
	customerRepo := db.NewCustomerRepository(pool)
	customerService := service.NewCustomerService(customerRepo)
	customerHandler := grpcserver.NewCustomerHandler(customerService)

	// create gRPC server — interceptors will be added here later
	grpcServer := grpc.NewServer()

	// register service handlers on the server
	pb.RegisterCustomerServiceServer(grpcServer, customerHandler)

	// reflection allows Evans and grpcurl to discover services without proto files
	if cfg.AppEnv == "development" {
		reflection.Register(grpcServer)
	}

	// start TCP listener on configured port
	listener, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.ServerPort, err)
	}

	fmt.Printf("gRPC server listening on port %s\n", cfg.ServerPort)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}