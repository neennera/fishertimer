# ACCOUNT Service Architecture

## Architecture Pattern: Hexagonal / Clean Architecture

This service is structured according to Ports & Adapters (Hexagonal) principles to ensure zero coupling between business logic and infrastructure.

```
account/
├── cmd/
│   └── main.go                    # Composition root (Dependency Injection)
├── config/
│   └── config.go                  # Configuration and environment variables
├── internal/
│   ├── domain/                    # Core Enterprise Logic (No external deps)
│   │   ├── entity.go              # Entities & Value Objects
│   │   └── repository.go          # Outbound Driven Port interface
│   ├── usecase/                   # Application Use Cases
│   │   ├── service.go             # Business operations orchestration
│   │   └── service_test.go        # Isolated unit tests with repository mocks
│   └── adapter/                   # Technical Adapters
│       ├── handler/               # Inbound Driving Adapter (HTTP / REST)
│       │   └── http_handler.go
│       └── repository/            # Outbound Driven Adapter (DB / InMemory)
│           └── memory_repo.go
```

## Rules for Extending This Service
1. **Never import `adapter` from `domain` or `usecase`**.
2. Define repository interfaces in `internal/domain/repository.go`.
3. Implement database adapters in `internal/adapter/repository/` (e.g. Supabase, MongoDB, Redis).
4. Run unit tests with `pnpm test`.
