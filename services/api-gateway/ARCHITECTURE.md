# API Gateway Service Architecture

## Architecture Pattern: Clean Routing Architecture

This service acts as the single entry point for the Client frontend website to access backend microservices:

```
services/api-gateway/
├── cmd/
│   └── main.go                    # Composition root (starts HTTP reverse proxy server)
├── config/
│   └── config.go                  # Configuration and service URL loader
└── internal/
    └── adapter/
        └── handler/
            ├── gateway_handler.go      # Reverse proxy routes and CORS middleware
            └── gateway_handler_test.go # Unit tests
```

## Responsibilities
1. Route client requests to target microservices without exposing internal ports directly to the browser.
2. Provide CORS headers (`Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`).
3. Handle health checks (`/health` and `/api/v1/gateway/status`).
