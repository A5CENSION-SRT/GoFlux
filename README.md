# GoFlux

A production-grade gRPC backend server built in Go, engineered to withstand 100,000 concurrent users. The system is modeled around a neobank payment platform, where banking infrastructure serves as the ideal domain for stress-testing—it combines write contention, complex queries, concurrent access patterns, and stringent correctness requirements.

## Project Philosophy

Rather than building a tutorial application that works "well enough," GoFlux emphasizes the engineering mindset: measure first, optimize second, prove with numbers. Every architectural decision is validated through load testing, query analysis, and observability metrics. The final deliverable is an engineering report demonstrating exactly how and why the system scales.

## What's Complete

### Database Layer
- 13-table PostgreSQL schema optimized for banking operations
- Monthly partitioned transactions table for horizontal scalability
- Deliberately engineered indexes on every high-traffic table
- 500,000+ rows of realistic seed data
- PgBouncer connection pooling via Docker Compose
- Migrations management via buf

### Protocol Layer
- 9 gRPC services defined
- 43 RPC methods across all services
- Protocol buffer code generation via buf
- 18 generated Go files in `internal/gen/`

### Project Structure
- `cmd/server/` — gRPC server entry point
- `cmd/worker/` — background job worker
- `cmd/seed/` — database initialization
- `internal/config/` — environment configuration
- `internal/db/` — repository layer with SQL queries
- `internal/service/` — business logic layer
- `internal/server/` — gRPC handler implementations
- `internal/middleware/` — interceptors and middleware
- `internal/models/` — domain models
- `proto/` — Protocol Buffer definitions

## Implementation Roadmap

### Phase 1: gRPC Server Implementation
Core gRPC server with layered architecture. Services implemented in dependency order:

1. **CustomerService** — Foundation layer, basic CRUD operations
2. **AccountService** — Depends on customers
3. **TransactionService** — Depends on accounts and merchants
4. **TransferService** — Multi-table transactions, complex business logic
5. **BudgetService** — Account-dependent analytics
6. **MerchantService** — Relatively independent
7. **AnalyticsService** — Complex window functions and CTEs
8. **SettlementService** — Batch processing visibility

### Phase 2: Observability Infrastructure
- Prometheus metrics exposed on `/metrics` endpoint
- Grafana dashboards tracking:
  - Per-RPC method latency (p50, p95, p99)
  - Database connection pool utilization
  - Cache hit rate
  - Error rate by method
  - System throughput over time
- Jaeger distributed tracing:
  - Full request path from handler to database
  - Query execution timing per layer
  - Latency attribution

### Phase 3: Load Testing Framework
- k6 scripts simulating realistic traffic distribution:
  - 40% GetAccountBalance
  - 25% ListAccountTransactions
  - 15% InitiateTransfer
  - 10% GetTransferStatus
  - 7% GetSpendingByCategory
  - 3% GetBudgetStatus

- Ramp profile progression:
  - 100 users (warmup)
  - 1,000 users (light load)
  - 10,000 users (medium load)
  - 50,000 users (heavy load)
  - 100,000 users (peak stress)

- ghz for isolated RPC benchmarking

### Phase 4: Query Optimization
- EXPLAIN ANALYZE on all slow queries
- Index addition based on query plans
- Monthly partitioning verification
- Before/after metrics at each optimization

### Phase 5: Connection Pool Tuning
- Incremental PgBouncer configuration adjustments
- Identifying throughput plateau
- Bottleneck analysis and documentation

### Phase 6: Caching Strategy
- Redis cache-aside for account balances (1-2s TTL)
- Merchant data caching (1h TTL)
- Category hierarchy caching (long TTL)
- Cache hit rate monitoring and latency analysis

### Phase 7: Security and Idempotency
- JWT interceptor with signature validation
- Token expiry checking
- Idempotency interceptor with duplicate request handling
- Correctness verification under concurrent load

### Phase 8: Background Jobs
- **Settlement processor** — End-of-day batch settlement
- **Scheduled payment processor** — Minute-level payment automation
- **Stuck transfer detector** — Stale transfer detection and compensation

## Performance Targets

The engineering report will document progression across the optimization phases:

**Baseline (Phase 1 Complete)**
- p99 latency: 820ms
- Throughput: 780 req/sec
- Error rate: 4.2%

**Post-Optimization Target (Phase 4-6)**
- p99 latency: <100ms
- Throughput: >2,500 req/sec
- Error rate: <0.1%

Each improvement is documented with specific changes and their measured impact.

## Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7.0+ (for caching phases)
- Docker and Docker Compose
- buf (for proto generation)
- k6 (for load testing)

### Local Development

1. Start the database and supporting services:
```bash
docker-compose up -d
```

2. Run migrations:
```bash
make migrate
```

3. Seed the database:
```bash
make seed
```

4. Build and run the server:
```bash
make build
./main
```

## Architecture

GoFlux follows a layered architecture ensuring clear separation of concerns:

```
gRPC Handlers (server/)
        ↓
Business Logic (service/)
        ↓
Data Access (db/)
        ↓
PostgreSQL
```

Each layer has a single responsibility:
- **Server**: Protocol translation and request validation
- **Service**: Business logic and orchestration
- **Database**: SQL queries and data persistence

## Technologies

- **Language**: Go 1.21
- **RPC Framework**: gRPC
- **Protocol**: Protocol Buffers v3
- **Database**: PostgreSQL 15
- **Connection Pooling**: pgxpool + PgBouncer
- **Caching**: Redis
- **Monitoring**: Prometheus + Grafana
- **Tracing**: Jaeger
- **Load Testing**: k6
- **Benchmarking**: ghz

## Learning Outcomes

Completing GoFlux provides expertise in:

- Go production patterns and best practices
- gRPC service design and protocol evolution
- PostgreSQL schema design, indexing strategies, and partitioning
- Connection pool management and exhaustion handling
- Distributed caching patterns and invalidation strategies
- Systems observability through metrics and tracing
- Load testing methodology and percentile analysis
- System design and scalability principles
- Engineering discipline: measuring, analyzing, and optimizing

## Documentation

- `rpc.md` — RPC method specifications and contracts
- Phase-specific analysis documents — In-depth optimization reports
- Performance baseline — Raw metrics before optimization
- Optimization deltas — Measured improvements at each phase

## License

This project is provided as-is for educational purposes.

---
