---
name: database-and-repository
description: Complete guide for AI agents on designing domain repositories, implementing database adapters, locating schemas, and understanding data models in Fisher Timer.
---

# Skill 4: Database & Repository Pattern Architecture

## Objective
Guide AI agents in implementing, extending, and testing repository patterns within Go microservices, adhering strictly to Clean/Hexagonal Architecture boundaries and locating schema files.

---

## 1. Where Are Schemas Located?

In our **Database-per-Service** architecture, each microservice owns its database configuration, migrations, and schemas inside its own directory:

```
services/<service-name>/database/
├── README.md                             # Service database documentation & port info
├── schemas/                              # Service-specific DDL / schema definitions (to be defined)
└── migrations/                           # Service-specific schema migration scripts
```

### Schema Convention:
- **Relational Databases (PostgreSQL):** Each service (`auth`, `account`, `study-session`, `study-timer`, `admin`) manages its own isolated database instance and schema.
- **Document Databases (MongoDB):** Each service (`reward`, `leaderboard`) manages its own isolated document database and collections.

---

## 2. Sample Data Models & Ownership

### Sample Table 1: `auth.users` & `account.profiles` (PostgreSQL)
- **Service Owners:** Auth Service & Account Service
- **File:** [`database/schemas/postgres/001_create_users_table.sql`](../../../database/schemas/postgres/001_create_users_table.sql)
- **Purpose:** Manages core user identity, Google OAuth metadata, personal focus statistics, and ban moderation status.

### Sample Table 2: `session.study_sessions` & `session_participants` (PostgreSQL)
- **Service Owner:** Study Session Service
- **File:** [`database/schemas/postgres/002_create_study_sessions_table.sql`](../../../database/schemas/postgres/002_create_study_sessions_table.sql)
- **Purpose:** Tracks room lifecycles (`ACTIVE`, `ENDED`), creator ownership, and atomic roster participant caps.

### Sample Collection 1: `fish_rewards` (MongoDB)
- **Service Owner:** Reward Service
- **File:** [`database/schemas/mongodb/001_create_fish_rewards_collection.js`](../../../database/schemas/mongodb/001_create_fish_rewards_collection.js)
- **Purpose:** Stores caught fish items with dynamic properties (rarity, species, weight score) and completion timestamps.

---

## 3. How to Create a Repository in Go Clean Architecture

In our Hexagonal architecture, repository development follows a two-part separation:

```mermaid
graph LR
    subgraph Domain Layer ["Core (No Dependencies)"]
        Interface["internal/domain/repository.go<br/>(Driven Port Interface)"]
    end

    subgraph Adapter Layer ["Technical Infrastructure"]
        Memory["internal/adapter/repository/memory_repo.go<br/>(Unit test mock)"]
        Postgres["internal/adapter/repository/postgres_repo.go<br/>(pgx / database/sql)"]
        Mongo["internal/adapter/repository/mongo_repo.go<br/>(mongo-go-driver)"]
    end

    Memory -.->|implements| Interface
    Postgres -.->|implements| Interface
    Mongo -.->|implements| Interface
```

### Step 1: Define the Repository Interface in `internal/domain/repository.go`
This is the **Driven Port**. It uses pure Go types and domain entities. It **must never** import SQL drivers, ORMs, or Mongo libraries.

```go
package domain

import "context"

type Repository interface {
    GetByID(ctx context.Context, id string) (*StudySession, error)
    Create(ctx context.Context, session *StudySession) error
}
```

### Step 2: Implement the Adapter in `internal/adapter/repository/`
This is the **Driven Adapter**. Create `postgres_repo.go` (or `mongo_repo.go`). It implements the methods declared in `domain.Repository`:

```go
package repository

import (
    "context"
    "database/sql"
    "github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type PostgresRepository struct {
    db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
    return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.StudySession, error) {
    query := `SELECT id, name, creator_id, participant_limit, status, created_at 
              FROM session.study_sessions WHERE id = $1`
    row := r.db.QueryRowContext(ctx, query, id)
    
    var s domain.StudySession
    if err := row.Scan(&s.ID, &s.Name, &s.CreatorID, &s.ParticipantLimit, &s.Status, &s.CreatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, domain.ErrNotFound
        }
        return nil, err
    }
    return &s, nil
}
```

### Step 3: Wire in `cmd/main.go` (Composition Root)
Switching between an in-memory repository for local tests and a PostgreSQL repository requires altering only `cmd/main.go`:

```go
// In cmd/main.go
db, err := sql.Open("postgres", cfg.DatabaseURL)
repo := repository.NewPostgres(db)
uc := usecase.New(repo)
h := handler.New(uc)
```

---

## 4. Connection Strings & Environment Variables

| Service | Environment Variable | Default Local Connection | Target Technology |
| :--- | :--- | :--- | :--- |
| `auth` | `DATABASE_URL` | `postgres://postgres:postgrespassword@localhost:5432/fishertimer?sslmode=disable` | Supabase / PostgreSQL |
| `account` | `DATABASE_URL` | `postgres://postgres:postgrespassword@localhost:5432/fishertimer?sslmode=disable` | Supabase / PostgreSQL |
| `study-session` | `DATABASE_URL` | `postgres://postgres:postgrespassword@localhost:5432/fishertimer?sslmode=disable` | Supabase / PostgreSQL |
| `study-timer` | `DATABASE_URL` | `postgres://postgres:postgrespassword@localhost:5432/fishertimer?sslmode=disable` | Supabase / PostgreSQL |
| `reward` | `MONGODB_URI` | `mongodb://mongoadmin:mongopassword@localhost:27017/fishertimer?authSource=admin` | MongoDB |
| `leaderboard`| `MONGODB_URI` | `mongodb://mongoadmin:mongopassword@localhost:27017/fishertimer?authSource=admin` | MongoDB |
| `admin` | `DATABASE_URL` | `postgres://postgres:postgrespassword@localhost:5432/fishertimer?sslmode=disable` | Supabase / PostgreSQL |
