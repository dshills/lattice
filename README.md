# Lattice

Lattice is a lightweight, graph-based work tracking system. It replaces heavyweight tools like Jira with a minimal interface: only three states, free-form tags, and first-class relationships between work items. Complexity lives in the data model and views, never in user-facing workflows.

## Key Concepts

- **Three states only:** `NotDone` -> `InProgress` -> `Completed`. Forward transitions are one step at a time. Any other transition (backward or skip) requires `override: true`.
- **Relationships are first-class:** `blocks`, `depends_on`, `relates_to`, `duplicate_of`. Relationships are directed, unique per (source, target, type), and support reverse lookup.
- **Tags are metadata:** Free-form strings that never affect state transitions or behavior.
- **Hierarchy via parent_id:** Each work item can have one parent. Max depth is 100 levels. Circular parent chains are rejected.
- **Cycles are allowed but detectable:** Circular dependencies via `depends_on`/`blocks` relationships are permitted. A dedicated endpoint detects them.
- **Derived status:** A work item is "blocked" if any `depends_on` target is not `Completed`. It is "ready" when all `depends_on` targets are `Completed`.

## Technology

- **Backend:** Go 1.23+, `net/http` with Go 1.22+ method-based routing, `database/sql` (no ORM)
- **Frontend:** React 19, TypeScript, Vite, Tailwind CSS, TanStack Query, React Flow, dnd-kit
- **Database:** MySQL 8.0+ (recursive CTEs required)

## Project Structure

```
lattice/
├── cmd/lattice/main.go           # Server entrypoint, config, graceful shutdown
├── internal/
│   ├── api/                      # HTTP handlers, middleware, error mapping
│   ├── domain/                   # State machine, validation, types
│   ├── graph/                    # DFS cycle detection with recursive CTE
│   └── store/mysql/              # MySQL CRUD, batch loading, migrations
├── frontend/
│   ├── src/
│   │   ├── app/                  # Providers, router, AppShell layout
│   │   ├── components/           # Reusable UI components
│   │   │   ├── common/           # Toast, Modal, LoadingState, ErrorState, EmptyState
│   │   │   ├── workitems/        # WorkItemCard, BoardColumn, StateSelector
│   │   │   ├── forms/            # CreateWorkItemForm, TagEditor, RelationshipEditor
│   │   │   ├── filters/          # FilterPanel, SearchInput
│   │   │   └── graph/            # GraphNode, GraphDetailPanel
│   │   ├── hooks/                # useWorkItems, useFilters, useRelationships, useCycles
│   │   ├── lib/                  # API client, types, validation (Zod), constants
│   │   └── pages/                # Home, Board, List, Graph, ItemDetail, Settings
│   ├── vite.config.ts
│   └── vitest.config.ts
├── migrations/                   # SQL migration files (up/down)
└── specs/                        # SPEC.md and PLAN.md
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- Node.js 20+ and npm
- MySQL 8.0 or later

### Database Setup

Create a MySQL database:

```sql
CREATE DATABASE lattice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'lattice'@'localhost' IDENTIFIED BY 'your-password';
GRANT ALL PRIVILEGES ON lattice.* TO 'lattice'@'localhost';
```

Migrations run automatically on server startup.

### Install Dependencies

```bash
make install    # runs go mod download + npm install
```

### Build and Run

```bash
# Build everything (Go binary + frontend)
make build

# Or use the Makefile targets for development:
make run-api    # Start API server with hot reload (air)
make run-ui     # Start Vite dev server on http://localhost:5175
```

The API server listens on `:8090` by default. The Vite dev server proxies `/workitems` requests to the API.

### Environment Variables

Create a `.env` file in the project root (loaded automatically by the Makefile):

```
LATTICE_DB_HOST=127.0.0.1
LATTICE_DB_PORT=3306
LATTICE_DB_USER=lattice
LATTICE_DB_PASSWORD=your-password
LATTICE_DB_NAME=lattice
LATTICE_ADDR=:8090
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LATTICE_DB_HOST` | Yes | — | MySQL host |
| `LATTICE_DB_PORT` | No | `3306` | MySQL port |
| `LATTICE_DB_USER` | Yes | — | MySQL user |
| `LATTICE_DB_PASSWORD` | No | — | MySQL password |
| `LATTICE_DB_NAME` | Yes | — | MySQL database name |
| `LATTICE_ADDR` | No | `:8080` | API listen address |
| `LATTICE_MIGRATIONS_DIR` | No | `migrations` | Path to SQL migration files |

### Database Migrations

```bash
make migrate          # Run all pending migrations
make migrate-down     # Roll back one migration
make migrate-status   # Show current version
```

Migrations also run automatically on API server startup.

### Running Tests

```bash
make test             # Run all tests (Go + frontend)
make test-go          # Go tests only
make test-frontend    # Frontend tests only (vitest)
make lint             # Lint everything (golangci-lint + eslint)
```

### All Makefile Targets

```bash
make help             # Show all available targets
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
| 403 | `FORBIDDEN` | Action not permitted |
| 404 | `NOT_FOUND` | Work item or relationship not found |
| 409 | `INVALID_TRANSITION` | State transition not allowed |
| 422 | `VALIDATION_ERROR` | Referential integrity, cycle, or depth violation |

### Endpoints

#### POST /workitems

Create a new work item. State is always set to `NotDone`. The `id`, `created_at`, and `updated_at` fields are system-generated.

```bash
curl -X POST http://localhost:8090/workitems \
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
curl http://localhost:8090/workitems/550e8400-e29b-41d4-a716-446655440000
```

**Response (200):** Full WorkItem object including tags and relationships.

#### PATCH /workitems/{id}

Partial update. Only provided fields are modified. Tags are replaced entirely if present. To unset `parent_id`, send `"parent_id": ""`.

```bash
curl -X PATCH http://localhost:8090/workitems/550e8400-... \
  -H "Content-Type: application/json" \
  -d '{"state": "InProgress"}'
```

For backward or skip transitions, include `override: true`:

```bash
curl -X PATCH http://localhost:8090/workitems/550e8400-... \
  -H "Content-Type: application/json" \
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
curl "http://localhost:8090/workitems?state=NotDone&tags=backend,urgent&page=1&page_size=20"
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
curl -X DELETE http://localhost:8090/workitems/550e8400-...
```

**Response:** 204 No Content.

#### POST /workitems/{id}/relationships

```bash
curl -X POST http://localhost:8090/workitems/550e8400-.../relationships \
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
curl -X DELETE http://localhost:8090/workitems/550e8400-.../relationships/770e8400-...
```

**Response:** 204 No Content.

#### GET /workitems/{id}/cycles

Detects dependency cycles (`depends_on` and `blocks` edges) involving the specified work item.

```bash
curl http://localhost:8090/workitems/550e8400-.../cycles
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
   ^             ^              │
   └─────────────┴── override ──┘
```

- Forward (one step): always allowed
- Backward or skip: requires `override: true` in the request body

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
