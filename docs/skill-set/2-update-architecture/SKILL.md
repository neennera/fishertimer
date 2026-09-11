---
name: update-architecture
description: Instructions and protocol for AI agents to automatically synchronize and maintain the system architecture documentation in /docs/ARCHITECTURE.md.
---

# Skill 2: Auto-Update System Architecture

## Objective
Prevent architectural drift by establishing a mandatory protocol for agents to keep [`/docs/ARCHITECTURE.md`](../../ARCHITECTURE.md) synchronized with the actual codebase.

---

## 1. Trigger Conditions
An agent MUST evaluate and update [`/docs/ARCHITECTURE.md`](../../ARCHITECTURE.md) whenever a task involves:
1. **Adding or Decommissioning a Service:** Modifying `services/*` or `apps/*`.
2. **Port or Protocol Changes:** Changing default listen ports, introducing WebSockets, gRPC, or messaging queues.
3. **Database or Persistence Modifications:** Introducing new tables, MongoDB collections, or altering storage boundaries between Supabase and Mongo.
4. **Inter-Service Communication:** Introducing a new REST call, RPC, or asynchronous event between two or more microservices.
5. **Layer or Dependency Inversions:** Updating the Clean/Hexagonal template within `internal/`.

---

## 2. Synchronization Checklist

When updating [`/docs/ARCHITECTURE.md`](../../ARCHITECTURE.md), check each of the following sections:

### Check 1: High-Level Mermaid Diagram (Section 1)
- If a new service or database was added, add the node in the `mermaid` code block.
- Update arrow edges (`-->` for direct HTTP, `-.->` for inter-service calls or events).

### Check 2: Microservice Topology & Port Registry (Section 2)
- Ensure the port, package name, directory path, protocol, and database mapping match the service's `config/config.go` and `cmd/main.go`.

### Check 3: Inter-Service Communication Matrix (Section 3)
- If `Service A` calls `Service B`, add or update the row in the table with the calling method, target endpoint, and trigger condition.

### Check 4: Shared Types & Contracts (Section 5)
- If new cross-cutting entities or DTOs were added to `@fishertimer/shared-types`, verify they are noted.

---

## 3. Heuristics for Clean Architecture Enforcement
When adding files to any service under `services/<name>/`:
- **Domain Layer (`internal/domain/`):** Must NEVER import external packages, HTTP routers, or SQL/Mongo drivers.
- **Usecase Layer (`internal/usecase/`):** Must orchestrate domain models using repository interfaces.
- **Adapter Layer (`internal/adapter/`):** Place HTTP/WebSocket handlers under `adapter/handler/` and database queries under `adapter/repository/`.
- If an agent detects a violation of these layer boundaries, refactor immediately and document in `CHANGELOG.md`.
