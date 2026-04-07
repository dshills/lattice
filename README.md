# Lattice

Lattice is a lightweight, graph-based work tracking system. It replaces heavyweight tools like Jira with a minimal interface: only three states, free-form tags, and first-class relationships between work items. Complexity lives in the data model and views, never in user-facing workflows.

## Key Concepts

- **Three states only:** `NotDone` -> `InProgress` -> `Completed`. Forward transitions are allowed one step at a time. Backward transitions require admin override.
- **Relationships are first-class:** `blocks`, `depends_on`, `relates_to`, `duplicate_of`. Relationships are directed, unique per (source, target, type), and support reverse lookup.
- **Tags are metadata:** Free-form strings that never affect state transitions or behavior.
- **Hierarchy via parent_id:** Each work item can have one parent. Max depth is 100 levels. Circular parent chains are rejected.
- **Cycles are allowed but detectable:** Circular dependencies via `depends_on`/`blocks` relationships are permitted. A dedicated endpoint detects them.
- **Derived status:** A work item is "blocked" if any `depends_on` target is not `Completed`. It is "ready" when all `depends_on` targets are `Completed`.

## Technology

- **Language:** Go 1.23+
- **Database:** MySQL 8.0+ (recursive CTEs required)
- **HTTP Router:** `net/http` with Go 1.22+ method-based routing
- **No ORM:** Uses `database/sql` with parameterized queries

## Project Structure

```
lattice/
├── cmd/lattice/main.go           # Server entrypoint, config, graceful shutdown
├── internal/
│   ├── api/
│   │   ├── errors.go             # Domain error -> HTTP status mapping
│   │   ├── handler.go            # All 8 HTTP handlers + route registration
│   │   ├── handler_test.go       # Handler unit tests (mock stores)
│   │   ├── integration_test.go   # End-to-end HTTP tests (requires MySQL)
│   │   └── middleware.go         # Logging, role extraction, content-type
│   ├── domain/
│   │   ├── errors.go             # Sentinel errors
│   │   ├── state.go              # State type, transitions, validation
│   │   └── workitem.go           # WorkItem struct, validation, constants
│   ├── graph/
│   │   └── cycle.go              # DFS cycle detection with recursive CTE
│   └── store/
│       ├── store.go              # Store interfaces (WorkItemStore, etc.)
│       └── mysql/
│           ├── migration.go      # SQL migration runner with advisory locking
│           ├── workitem.go       # WorkItem CRUD, batch loading, hierarchy
│           └── relationship.go   # Relationship Add/Remove/ListByTarget
├── migrations/
│   ├── 001_create_work_items.{up,down}.sql
│   ├── 002_create_work_item_tags.{up,down}.sql
│   └── 003_create_work_item_relationships.{up,down}.sql
└── specs/initial/
    ├── SPEC.md                   # Full system specification
    └── PLAN.md                   # Phased implementation plan
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- MySQL 8.0 or later

### Database Setup

Create a MySQL database:

```sql
CREATE DATABASE lattice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'lattice'@'localhost' IDENTIFIED BY 'your-password';
GRANT ALL PRIVILEGES ON lattice.* TO 'lattice'@'localhost';
```

Migrations run automatically on server startup.

### Build and Run

```bash
# Build
go build -o lattice ./cmd/lattice

# Run (migrations apply automatically)
LATTICE_DSN="lattice:your-password@tcp(127.0.0.1:3306)/lattice?parseTime=true&multiStatements=true" \
  ./lattice
```

The server starts on `:8080` by default.

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LATTICE_DSN` | Yes | — | MySQL DSN (must include `parseTime=true&multiStatements=true`) |
| `LATTICE_ADDR` | No | `:8080` | Listen address |
| `LATTICE_MIGRATIONS_DIR` | No | `migrations` | Path to SQL migration files |

### Running Tests

```bash
# Unit tests (no database required)
go test ./...

# Integration tests (requires MySQL)
LATTICE_TEST_DSN="lattice:your-password@tcp(127.0.0.1:3306)/lattice_test?parseTime=true&multiStatements=true" \
  go test ./... -count=1

# Single test
go test -run TestName ./path/to/package

# Lint
golangci-lint run ./...
```

## API Reference

All request/response bodies use `application/json`. Errors follow a consistent format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "not found: work item \"abc\""
  }
}
```

### Error Codes

| HTTP Status | Code | Cause |
|-------------|------|-------|
| 400 | `INVALID_INPUT` | Malformed request, field constraint violation |
| 403 | `FORBIDDEN` | Override requested without admin role |
| 404 | `NOT_FOUND` | Work item or relationship not found |
| 409 | `INVALID_TRANSITION` | State transition not allowed |
| 422 | `VALIDATION_ERROR` | Referential integrity, cycle, or depth violation |

### Endpoints

#### POST /workitems

Create a new work item. State is always set to `NotDone`. The `id`, `created_at`, and `updated_at` fields are system-generated.

```bash
curl -X POST http://localhost:8080/workitems \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Implement auth",
    "description": "Add JWT-based authentication",
    "type": "feature",
    "tags": ["backend", "security"],
    "parent_id": null
  }'
```

**Response (201):** Full WorkItem object.

#### GET /workitems/{id}

```bash
curl http://localhost:8080/workitems/550e8400-e29b-41d4-a716-446655440000
```

**Response (200):** Full WorkItem object including tags and relationships.

#### PATCH /workitems/{id}

Partial update. Only provided fields are modified. Tags are replaced entirely if present. To unset `parent_id`, send `"parent_id": ""`.

```bash
curl -X PATCH http://localhost:8080/workitems/550e8400-... \
  -H "Content-Type: application/json" \
  -d '{"state": "InProgress"}'
```

For backward state transitions, include `override: true` and set the `X-Role: admin` header:

```bash
curl -X PATCH http://localhost:8080/workitems/550e8400-... \
  -H "Content-Type: application/json" \
  -H "X-Role: admin" \
  -d '{"state": "NotDone", "override": true}'
```

**Response (200):** Full updated WorkItem object.

#### GET /workitems

List work items with filtering and pagination.

| Parameter | Type | Description |
|-----------|------|-------------|
| `state` | string | `NotDone`, `InProgress`, or `Completed` |
| `tags` | string | Comma-separated; AND logic (all must match) |
| `type` | string | Filter by type value |
| `parent_id` | UUID | Filter by parent |
| `relationship_type` | string | `blocks`, `depends_on`, `relates_to`, `duplicate_of` |
| `relationship_target_id` | UUID | Combined with `relationship_type` |
| `is_blocked` | bool | Has unresolved `depends_on` |
| `is_ready` | bool | All `depends_on` targets are `Completed` |
| `page` | int | Default: 1 |
| `page_size` | int | Default: 50, max: 200 |

```bash
curl "http://localhost:8080/workitems?state=NotDone&tags=backend,urgent&page=1&page_size=20"
```

**Response (200):**

```json
{
  "items": [...],
  "total": 42,
  "page": 1,
  "page_size": 20
}
```

#### DELETE /workitems/{id}

Deletes a work item atomically. Cascades: removes all relationships (both directions), nulls `parent_id` on children, removes tags.

```bash
curl -X DELETE http://localhost:8080/workitems/550e8400-...
```

**Response:** 204 No Content.

#### POST /workitems/{id}/relationships

```bash
curl -X POST http://localhost:8080/workitems/550e8400-.../relationships \
  -H "Content-Type: application/json" \
  -d '{"type": "depends_on", "target_id": "660e8400-..."}'
```

**Response (201):**

```json
{
  "id": "generated-uuid",
  "type": "depends_on",
  "target_id": "660e8400-..."
}
```

#### DELETE /workitems/{id}/relationships/{rel_id}

```bash
curl -X DELETE http://localhost:8080/workitems/550e8400-.../relationships/770e8400-...
```

**Response:** 204 No Content.

#### GET /workitems/{id}/cycles

Detects dependency cycles (`depends_on` and `blocks` edges) involving the specified work item.

```bash
curl http://localhost:8080/workitems/550e8400-.../cycles
```

**Response (200):**

```json
{
  "cycles": [
    ["550e8400-...", "660e8400-..."]
  ]
}
```

Returns an empty array if no cycles exist.

## Data Model

### WorkItem

```json
{
  "id": "UUID v4 (system-generated)",
  "title": "string (required, max 500 chars)",
  "description": "string (max 10000 chars)",
  "state": "NotDone | InProgress | Completed",
  "tags": ["string (max 100 chars each, no commas, max 50)"],
  "type": "string (optional, max 100 chars)",
  "parent_id": "UUID v4 or null",
  "relationships": [
    {
      "id": "UUID v4",
      "type": "blocks | depends_on | relates_to | duplicate_of",
      "target_id": "UUID v4"
    }
  ],
  "created_at": "2026-04-07T12:00:00Z (immutable)",
  "updated_at": "2026-04-07T12:00:00Z (auto-updated)"
}
```

### State Machine

```
NotDone ──> InProgress ──> Completed
   ^                          │
   └──── admin override ──────┘
```

- Forward: always allowed (one step at a time)
- Skip forward (NotDone -> Completed): not allowed
- Backward: requires `override: true` in request body + `X-Role: admin` header

### Database Schema

Three tables with foreign keys and cascading deletes:

- **work_items** — Core work item data with indexes on state, type, parent_id
- **work_item_tags** — Composite PK (item_id, tag), FK to work_items with CASCADE
- **work_item_relationships** — Unique constraint on (source_id, target_id, type), FKs with CASCADE

Migrations are applied automatically on startup using advisory locking to prevent concurrent execution.

## Architecture Notes

- **Store layer** uses batch loading (3 queries total for List, not N+1) and recursive CTEs for hierarchy traversal
- **Cycle detection** loads the reachable subgraph via recursive CTE in a single query, then runs DFS in memory
- **Migration runner** uses MySQL advisory locks (`GET_LOCK`/`RELEASE_LOCK`) on a dedicated connection to prevent concurrent migrations
- **No transactions for DDL** — MySQL implicitly commits DDL statements, so each migration file contains a single DDL operation
- The `X-Role` header is trusted as-is (designed to sit behind an API gateway that sets it)
