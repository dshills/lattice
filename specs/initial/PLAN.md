# PLAN.md

# Lattice — Phased Implementation Plan

## Technology Stack

- **Language:** Go 1.23+
- **Database:** MySQL 8.0+ (accessed via `database/sql` + `github.com/go-sql-driver/mysql`)
- **HTTP Router:** `net/http` with Go 1.22+ routing patterns (method + path params)
- **ID Generation:** `github.com/google/uuid`
- **Testing:** `testing` stdlib + `github.com/stretchr/testify` for assertions
- **Linting:** `golangci-lint`
- **Migration:** SQL files applied at startup or via CLI flag; each migration has a corresponding `down` file for rollback

## Project Layout

```
lattice/
├── cmd/
│   └── lattice/
│       └── main.go              # entrypoint, config, server startup
├── internal/
│   ├── domain/
│   │   ├── workitem.go          # WorkItem, State, Relationship, Tag types
│   │   ├── workitem_test.go
│   │   ├── state.go             # state transition logic + override
│   │   └── state_test.go
│   ├── store/
│   │   ├── mysql/
│   │   │   ├── workitem.go      # MySQL WorkItemStore implementation
│   │   │   ├── workitem_test.go # integration tests
│   │   │   ├── relationship.go  # MySQL RelationshipStore implementation
│   │   │   ├── relationship_test.go
│   │   │   └── migration.go     # schema migration runner
│   │   └── store.go             # Store interfaces
│   ├── api/
│   │   ├── handler.go           # HTTP handlers
│   │   ├── handler_test.go
│   │   ├── middleware.go         # role extraction from X-Role header
│   │   ├── errors.go            # error response helpers
│   │   └── validation.go        # request validation
│   └── graph/
│       ├── cycle.go             # cycle detection algorithm
│       └── cycle_test.go
├── migrations/
│   ├── 001_create_work_items.up.sql
│   ├── 001_create_work_items.down.sql
│   ├── 002_create_work_item_tags.up.sql
│   ├── 002_create_work_item_tags.down.sql
│   ├── 003_create_work_item_relationships.up.sql
│   └── 003_create_work_item_relationships.down.sql
├── specs/
│   └── initial/
│       ├── SPEC.md
│       └── PLAN.md
├── go.mod
└── go.sum
```

## Store Interfaces

Defined in `internal/store/store.go`. All methods accept `context.Context` as the first parameter.

```go
type WorkItemStore interface {
    Create(ctx context.Context, item *domain.WorkItem) error
    Get(ctx context.Context, id string) (*domain.WorkItem, error)
    Update(ctx context.Context, item *domain.WorkItem) error
    Delete(ctx context.Context, id string) error  // atomic: cascades relationships, nulls children
    List(ctx context.Context, filter ListFilter) (*ListResult, error)
    AncestorDepth(ctx context.Context, parentID string) (int, error)  // for hierarchy depth check
    HasCycle(ctx context.Context, childID, parentID string) (bool, error)  // parent cycle check
}

type RelationshipStore interface {
    Add(ctx context.Context, ownerID string, rel *domain.Relationship) error
    Remove(ctx context.Context, ownerID, relID string) error
    ListByTarget(ctx context.Context, targetID string) ([]domain.Relationship, error)
}

// CycleDetector detects dependency graph cycles (depends_on + blocks edges).
// Implemented by internal/graph/cycle.go; used by the GET /workitems/{id}/cycles handler.
// Distinct from WorkItemStore.HasCycle, which checks parent-child hierarchy cycles only.
type CycleDetector interface {
    DetectCycles(ctx context.Context, workItemID string) ([][]string, error)
}
```

---

## Phase 1: Domain Model + Database Schema

**Goal:** Establish the foundational types, validation logic, and database schema.

### Tasks

1. **Initialize Go module and dependencies**
   - `go mod init github.com/dshills/lattice`
   - Add dependencies:
     - `github.com/go-sql-driver/mysql` — standard, most widely used MySQL driver for Go's `database/sql`
     - `github.com/google/uuid` — well-maintained UUID v4 generation from Google
     - `github.com/stretchr/testify` — dominant Go assertion/mock library, avoids verbose manual comparisons

2. **Define domain types** (`internal/domain/workitem.go`)
   - `State` type with constants: `NotDone`, `InProgress`, `Completed`
   - `RelationshipType` type with constants: `Blocks`, `DependsOn`, `RelatesTo`, `DuplicateOf`
   - `Relationship` struct: `ID`, `Type`, `TargetID`
   - `WorkItem` struct: all fields from spec (ID, Title, Description, State, Tags, Type, ParentID, Relationships, CreatedAt, UpdatedAt)

3. **Implement state transition logic** (`internal/domain/state.go`)
   - `ValidateTransition(current, next State, override bool, isAdmin bool) error`
   - Forward transitions: NotDone→InProgress, InProgress→Completed (always allowed)
   - Backward transitions: only with override=true AND isAdmin=true
   - Returns typed errors: `ErrInvalidTransition`, `ErrForbidden`

4. **Implement field validation** (`internal/domain/workitem.go`)
   - `Validate() error` method on WorkItem for create/update
   - Title: required, max 500 chars
   - Description: max 10000 chars
   - Type: max 100 chars
   - Tags: each max 100 chars, no commas, max 50 tags
   - ParentID: not equal to own ID

5. **Create SQL migrations** (`migrations/`)
   - `001_create_work_items.up.sql`: work_items table with CHAR(36) PKs, appropriate column types, indexes on state, type, parent_id
   - `001_create_work_items.down.sql`: DROP TABLE work_items
   - `002_create_work_item_tags.up.sql`: work_item_tags table with composite PK (item_id, tag), FK to work_items
   - `002_create_work_item_tags.down.sql`: DROP TABLE work_item_tags
   - `003_create_work_item_relationships.up.sql`: work_item_relationships table with CHAR(36) PK, unique constraint on (source_id, target_id, type), FKs to work_items
   - `003_create_work_item_relationships.down.sql`: DROP TABLE work_item_relationships

6. **Write migration runner** (`internal/store/mysql/migration.go`)
   - Reads and applies SQL files in order (`.up.sql` for forward, `.down.sql` for rollback)
   - Tracks applied migrations in a `schema_migrations` table
   - Supports `migrate up` (apply all pending) and `migrate down N` (rollback N steps)

7. **Unit tests for domain logic**
   - State transitions: all valid forward, all invalid backward, backward with override (admin/non-admin)
   - Field validation: boundary cases for all constraints
   - Tag validation: comma rejection, length limits, count limits

### Exit Criteria
- All domain types compile and are tested
- Migrations can be applied to a fresh MySQL database
- `golangci-lint run ./...` passes

---

## Phase 2: WorkItem CRUD Store Layer

**Goal:** Implement MySQL-backed WorkItem persistence with full CRUD.

### Tasks

1. **Implement `WorkItemStore.Create`** (`internal/store/mysql/workitem.go`)
   - Initialize state to `domain.NotDone` (cannot be overridden by caller)
   - Generate UUID v4 for ID, set created_at/updated_at to now (UTC). The store layer ignores any caller-supplied values for id, created_at, and updated_at — these fields are always system-generated. (API layer rejects them with 400; the store is a second line of defense.)
   - Insert into work_items + bulk insert into work_item_tags
   - Validate parent_id references an existing WorkItem (if non-null)
   - Single transaction

2. **Implement `WorkItemStore.Get`**
   - Join work_items with work_item_tags and work_item_relationships
   - Assemble full WorkItem struct with tags and relationships arrays

3. **Implement `WorkItemStore.Update`**
   - Partial update: only modify provided fields
   - Replace tags entirely (delete all + re-insert)
   - Validate parent_id if changed
   - Always overwrite updated_at with current UTC time (ignores any caller-supplied value)
   - The store never modifies id or created_at
   - Single transaction

4. **Implement `WorkItemStore.Delete`**
   - Atomic transaction:
     1. Delete from work_item_relationships where source_id = id
     2. Delete from work_item_relationships where target_id = id
     3. Update work_items SET parent_id = NULL WHERE parent_id = id
     4. Delete from work_item_tags where item_id = id
     5. Delete from work_items where id = id
   - Rollback on any failure

5. **Implement `WorkItemStore.List`**
   - Dynamic query builder for filters: state, tags (AND logic), type, parent_id
   - Pagination: page + page_size with LIMIT/OFFSET
   - COUNT query for total
   - Returns `ListResult{Items, Total, Page, PageSize}`

6. **Implement `WorkItemStore.AncestorDepth` and `HasCycle`**
   - `AncestorDepth`: walk parent_id chain up, return depth count (for 100-level limit)
   - `HasCycle`: walk parent_id chain from parentID, check if childID is encountered

7. **Integration tests**
   - Require a test MySQL database (skip if unavailable)
   - Test all CRUD operations, cascade delete, orphan parent nullification
   - Test list filters individually and in combination
   - Test parent depth and cycle detection

### Exit Criteria
- All store methods pass integration tests against MySQL
- Cascade delete is atomic and tested
- List filtering works for all query params
- `golangci-lint run ./...` passes

---

## Phase 3: Relationship Store + Cycle Detection

**Goal:** Implement relationship management and cycle detection algorithm.

### Tasks

1. **Implement `RelationshipStore.Add`** (`internal/store/mysql/relationship.go`)
   - Generate UUID v4 for relationship ID
   - Validate: source WorkItem exists (404), target WorkItem exists (422)
   - Validate: type is one of the four allowed values
   - Enforce uniqueness constraint (source_id, target_id, type) — handle MySQL duplicate key error as 422
   - Reject system-generated fields in request

2. **Implement `RelationshipStore.Remove`**
   - Verify relationship belongs to specified WorkItem
   - Return 404 if WorkItem or relationship not found

3. **Implement `RelationshipStore.ListByTarget`**
   - Query relationships where target_id matches — supports reverse lookup

4. **Implement cycle detection** (`internal/graph/cycle.go`)
   - Algorithm: DFS from the specified WorkItem following `depends_on` and `blocks` edges (both as directed A→B)
   - Returns all unique cycles that include the starting WorkItem
   - Each cycle is a list of WorkItem IDs in traversal order
   - Load the subgraph lazily: query relationships from DB as needed during traversal (or batch-load if graph is bounded)

5. **Extend `WorkItemStore.List`** for relationship filters
   - `relationship_type`: JOIN work_item_relationships, filter by type
   - `relationship_target_id`: JOIN work_item_relationships, filter by target_id
   - When both provided: single JOIN with both conditions on same row

6. **Extend `WorkItemStore.List`** for derived filters
   - `is_blocked=true`: WorkItems with at least one `depends_on` relationship where target.state != 'Completed'
   - `is_blocked=false`: inverse
   - `is_ready=true`: WorkItems where ALL `depends_on` targets are 'Completed' (including those with none)
   - `is_ready=false`: inverse
   - Implemented via subqueries or LEFT JOINs

7. **Unit tests for cycle detection**
   - No cycles → empty result
   - Simple 2-node cycle
   - Complex multi-node cycle
   - Self-referential relationship
   - Mixed depends_on and blocks edges
   - WorkItem in multiple cycles

8. **Integration tests for relationship store**
   - Add/remove relationships
   - Duplicate rejection
   - Referential integrity (missing source/target)
   - Reverse lookup

### Exit Criteria
- Relationships can be added, removed, and queried
- Cycle detection correctly identifies all cycles through a WorkItem
- `is_blocked` and `is_ready` filters return correct results
- `golangci-lint run ./...` passes

---

## Phase 4: HTTP API Layer

**Goal:** Expose all functionality as a REST API with full validation and error handling.

### Tasks

1. **Implement error response helpers** (`internal/api/errors.go`)
   - `WriteError(w, status, code, message)` — writes JSON error response
   - Map domain errors to HTTP status codes:
     - `ErrNotFound` → 404 NOT_FOUND
     - `ErrInvalidTransition` → 409 INVALID_TRANSITION
     - `ErrForbidden` → 403 FORBIDDEN
     - `ErrValidation` → 422 VALIDATION_ERROR
     - `ErrInvalidInput` → 400 INVALID_INPUT

2. **Implement request validation** (`internal/api/validation.go`)
   - Parse and validate JSON request bodies
   - Reject system-generated fields (id, created_at, updated_at)
   - Validate field lengths, tag constraints, pagination params
   - Parse query parameters with type checking

3. **Implement middleware** (`internal/api/middleware.go`)
   - Role extraction: read first `X-Role` header, determine isAdmin bool
   - Request logging (method, path, status, duration)
   - Content-Type enforcement (application/json)

4. **Implement handlers** (`internal/api/handler.go`)
   - `POST /workitems` — Create WorkItem
   - `PATCH /workitems/{id}` — Update WorkItem (with state transition + override logic)
   - `GET /workitems/{id}` — Get WorkItem
   - `GET /workitems` — List WorkItems (all query params)
   - `DELETE /workitems/{id}` — Delete WorkItem
   - `POST /workitems/{id}/relationships` — Add Relationship
   - `DELETE /workitems/{id}/relationships/{rel_id}` — Remove Relationship
   - `GET /workitems/{id}/cycles` — Detect Cycles

5. **Wire up router** (`cmd/lattice/main.go`)
   - Parse config (DB DSN, listen address) from env vars
   - Initialize DB connection, run migrations
   - Register routes with `http.ServeMux`
   - Graceful shutdown on SIGINT/SIGTERM

6. **Handler tests**
   - Test each endpoint with valid and invalid inputs
   - Test error responses match spec (status code, error code, message)
   - Test state transition scenarios including override + admin role
   - Test pagination edge cases (page 0, negative, page_size > 200)
   - Test all query param filters
   - Use an interface-based mock store for unit tests

### Exit Criteria
- All 8 API endpoints respond correctly per spec
- Error responses match the specified format and codes
- Input validation rejects all invalid inputs per spec
- Admin override works correctly via X-Role header
- `golangci-lint run ./...` passes

---

## Phase 5: Hierarchy Enforcement + Integration Tests

**Goal:** Implement hierarchy depth limits, cycle prevention, and end-to-end integration tests.

### Tasks

1. **Enforce hierarchy depth limit on Create/Update**
   - Before setting parent_id, call `AncestorDepth(parentID)` and verify total depth < 100
   - Return `422 VALIDATION_ERROR` if exceeded

2. **Enforce parent cycle prevention on Update**
   - Before updating parent_id, call `HasCycle(workItemID, newParentID)`
   - Return `422 VALIDATION_ERROR` if cycle detected

3. **End-to-end integration tests**
   - Start a real MySQL instance (docker-compose or testcontainers)
   - Start the HTTP server
   - Test full request/response flows through HTTP:
     - Create → Get → Update → List → Delete lifecycle
     - Relationship CRUD with referential integrity checks
     - Cascade delete verification (relationships removed, children orphaned)
     - State transitions: forward, backward (admin/non-admin), override
     - Hierarchy: create deep chains, reject at 100 levels, reject cycles
     - Cycle detection with complex graphs
     - All filter combinations on List
     - Pagination correctness
     - All error scenarios per spec

4. **Verify acceptance criteria**
   - Map each of the 9 acceptance criteria to specific integration tests
   - Ensure all pass

### Exit Criteria
- Hierarchy depth limit enforced at 100 levels
- Parent-child cycles are detected and rejected
- All 9 acceptance criteria verified by integration tests
- Full API contract matches spec
- `golangci-lint run ./...` passes

---

## Cross-Cutting Concerns

These apply across all phases:

- **Transactions:** All multi-table mutations use database transactions with rollback on error
- **UUID generation:** Use `github.com/google/uuid` v4 for all IDs
- **Timestamps:** Always UTC, ISO 8601 format in JSON responses
- **Error mapping:** Domain errors → HTTP errors is centralized in the API layer
- **No ORM:** Use `database/sql` directly with parameterized queries to avoid SQL injection
- **Context propagation:** All store methods accept `context.Context` for cancellation/timeout
- **Lint:** Run `golangci-lint run ./...` after every phase

## Risk Mitigations

| Risk | Mitigation |
|------|------------|
| Cycle detection performance on large graphs | Lazy DFS with visited-set; bound traversal depth; consider caching in future |
| Hierarchy depth check is O(depth) per write | Acceptable for max 100; could denormalize depth column later if needed |
| `is_blocked`/`is_ready` filters require subqueries | Use indexed joins; monitor query plans; add composite indexes if slow |
| Atomic delete touches multiple tables | Single transaction with explicit ordering; tested in integration suite |
