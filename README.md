# Lattice

Lattice is a lightweight, graph-based work tracking system. It replaces heavyweight tools like Jira with a minimal interface: only three states, free-form tags, and first-class relationships between work items. Complexity lives in the data model and views, never in user-facing workflows.

## Key Concepts

- **Three states only:** `NotDone` -> `InProgress` -> `Completed`. Forward transitions are one step at a time. Any other transition (backward or skip) requires `override: true`.
- **Relationships are first-class:** `blocks`, `depends_on`, `relates_to`, `duplicate_of`. Relationships are directed, unique per (source, target, type), and support reverse lookup.
- **Tags are metadata:** Free-form strings that never affect state transitions or behavior.
- **Hierarchy via parent_id:** Each work item can have one parent. Max depth is 100 levels. Circular parent chains are rejected.
- **Cycles are allowed but detectable:** Circular dependencies via `depends_on`/`blocks` relationships are permitted. A dedicated endpoint detects them.
- **Derived status:** A work item is "blocked" if any `depends_on` target is not `Completed`. It is "ready" when all `depends_on` targets are `Completed`.
- **Multi-user with project roles:** Users authenticate via JWT. Each project has members with one of three roles: `owner`, `member`, or `viewer`. Owners manage members and project settings. Members create and edit work items. Viewers have read-only access.

## Technology

- **Backend:** Go 1.23+, `net/http` with Go 1.22+ method-based routing, `database/sql` (no ORM)
- **Frontend:** React 19, TypeScript, Vite, Tailwind CSS, TanStack Query, React Flow, dnd-kit
- **Database:** MySQL 8.0+ (recursive CTEs required)
- **Authentication:** JWT (HS256) with access/refresh token pair; refresh token stored in HttpOnly cookie

## Project Structure

```
lattice/
├── cmd/lattice/main.go           # Server entrypoint, config, graceful shutdown
├── internal/
│   ├── api/                      # HTTP handlers, middleware, error mapping
│   │   ├── handler.go            # Route registration, project & work item handlers
│   │   ├── auth_handler.go       # Register, login, refresh endpoints
│   │   ├── auth_middleware.go     # JWT validation middleware
│   │   ├── member_handler.go     # Project member management endpoints
│   │   ├── user_handler.go       # User profile endpoints (GET/PATCH /users/me)
│   │   ├── project_role_middleware.go  # Role extraction from project membership
│   │   └── authz.go              # Authorization helpers (requireOwner, requireWriteAccess)
│   ├── auth/                     # JWT token generation/validation, password hashing
│   ├── domain/                   # State machine, validation, types
│   │   ├── workitem.go           # WorkItem entity and validation
│   │   ├── user.go               # User entity and validation
│   │   ├── membership.go         # ProjectRole, ProjectMembership
│   │   └── state.go              # State transition rules
│   ├── graph/                    # DFS cycle detection with recursive CTE
│   └── store/mysql/              # MySQL CRUD, batch loading, migrations
│       ├── workitem.go           # Work item store with assignee JOIN
│       ├── user.go               # User store with bcrypt password hashing
│       ├── membership.go         # Membership store with role management
│       └── project.go            # Project store
├── frontend/
│   ├── src/
│   │   ├── app/                  # Providers, router, AppShell layout
│   │   ├── components/           # Reusable UI components
│   │   │   ├── common/           # Toast, Modal, LoadingState, ErrorState, EmptyState
│   │   │   ├── workitems/        # WorkItemCard, BoardColumn, StateSelector, AssigneeSelector
│   │   │   ├── forms/            # CreateWorkItemForm, TagEditor, RelationshipEditor
│   │   │   ├── filters/          # FilterPanel, SearchInput
│   │   │   └── graph/            # GraphNode, GraphDetailPanel
│   │   ├── hooks/                # useWorkItems, useAuth, useProjectRole, useMembers, useFilters
│   │   ├── lib/                  # API client, types, auth token management, constants
│   │   │   └── api/              # Per-resource API modules (auth, projects, workitems, members)
│   │   └── pages/                # Home, Board, List, Graph, ItemDetail, Members, Login, Register, Settings
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

The API server listens on `:8090` by default. The Vite dev server proxies `/projects`, `/auth`, and `/users` requests to the API.

### Environment Variables

Create a `.env` file in the project root (loaded automatically by the Makefile):

```
LATTICE_DB_HOST=127.0.0.1
LATTICE_DB_PORT=3306
LATTICE_DB_USER=lattice
LATTICE_DB_PASSWORD=your-password
LATTICE_DB_NAME=lattice
LATTICE_ADDR=:8090
LATTICE_JWT_SECRET=your-secret-key-at-least-32-characters
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LATTICE_DB_HOST` | Yes | — | MySQL host |
| `LATTICE_DB_PORT` | No | `3306` | MySQL port |
| `LATTICE_DB_USER` | Yes | — | MySQL user |
| `LATTICE_DB_PASSWORD` | No | — | MySQL password |
| `LATTICE_DB_NAME` | Yes | — | MySQL database name |
| `LATTICE_ADDR` | No | `:8090` | API listen address |
| `LATTICE_MIGRATIONS_DIR` | No | `migrations` | Path to SQL migration files |
| `LATTICE_JWT_SECRET` | Yes | — | HS256 signing key (min 32 characters) |
| `LATTICE_ACCESS_TOKEN_TTL` | No | `15m` | Access token lifetime |
| `LATTICE_REFRESH_TOKEN_TTL` | No | `168h` | Refresh token lifetime (default 7 days) |

### Database Migrations

```bash
make migrate          # Run all pending migrations
make migrate-down     # Roll back one migration
make migrate-status   # Show current version
```

Migrations also run automatically on API server startup. The current migrations are:

1. `001_create_work_items` — Core work items table
2. `002_create_work_item_tags` — Tags table
3. `003_create_work_item_relationships` — Relationships table
4. `004_add_target_id_index` — Index on relationship target_id
5. `005_add_projects` — Projects table, project_id FK on work items
6. `006_add_users` — Users table, project_memberships table, assignee_id/created_by on work items

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

All request/response bodies use `application/json`. Authenticated endpoints require a Bearer token in the `Authorization` header. Errors follow a consistent format:

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
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Insufficient role for the action |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `INVALID_TRANSITION` | State transition not allowed |
| 409 | `CONFLICT` | Duplicate email on registration |
| 422 | `VALIDATION_ERROR` | Referential integrity, cycle, or depth violation |

### Authentication Endpoints

#### POST /auth/register

Create a new account. Returns access token and sets refresh token cookie.

```bash
curl -X POST http://localhost:8090/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "display_name": "Jane Doe",
    "password": "securepassword"
  }'
```

**Response (201):** `{ "user": User, "access_token": "..." }`

#### POST /auth/login

```bash
curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{ "email": "user@example.com", "password": "securepassword" }'
```

**Response (200):** `{ "user": User, "access_token": "..." }`

#### POST /auth/refresh

Exchange refresh token cookie for a new access token.

**Response (200):** `{ "access_token": "..." }`

### User Endpoints

#### GET /users/me

Returns the authenticated user's profile.

#### PATCH /users/me

Update display name or password: `{ "display_name": "...", "password": "..." }`

### Project Endpoints

#### POST /projects

Create a project. The creator is automatically added as `owner`.

```bash
curl -X POST http://localhost:8090/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "name": "My Project", "description": "Optional description" }'
```

**Response (201):** Project object.

#### GET /projects

List projects the authenticated user is a member of. Includes the user's role per project.

**Response (200):** `{ "projects": [{ ...project, "item_count": 5, "role": "owner" }] }`

#### GET /projects/{project_id}

**Response (200):** Project object.

#### PATCH /projects/{project_id}

Update name or description. **Owner only.**

#### DELETE /projects/{project_id}

**Owner only.** Response: 204 No Content.

### Member Endpoints

All scoped to `/projects/{project_id}/members`.

#### GET /projects/{project_id}/members

List all members with their roles. Any project member can view.

**Response (200):** `{ "members": [{ "user_id", "email", "display_name", "role", ... }] }`

#### POST /projects/{project_id}/members

Add a member by email. **Owner only.**

```bash
curl -X POST http://localhost:8090/projects/$PID/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "email": "colleague@example.com", "role": "member" }'
```

#### PATCH /projects/{project_id}/members/{user_id}

Change a member's role. **Owner only.** `{ "role": "viewer" }`

#### DELETE /projects/{project_id}/members/{user_id}

Remove a member. **Owner only.** Cannot remove the last owner.

### Work Item Endpoints

All scoped to `/projects/{project_id}/workitems`. Require at least `viewer` role. Creating/updating/deleting require `member` or `owner` role.

#### POST /projects/{project_id}/workitems

Create a new work item. State is always set to `NotDone`. The `created_by` field is set to the authenticated user.

```bash
curl -X POST http://localhost:8090/projects/$PID/workitems \
  -H "Authorization: Bearer $TOKEN" \
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

#### GET /projects/{project_id}/workitems/{id}

**Response (200):** Full WorkItem object including tags, relationships, and assignee name.

#### PATCH /projects/{project_id}/workitems/{id}

Partial update. Only provided fields are modified. Tags are replaced entirely if present. To unset `parent_id` or `assignee_id`, send an empty string.

```bash
curl -X PATCH http://localhost:8090/projects/$PID/workitems/$ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"state": "InProgress"}'
```

For backward or skip transitions, include `override: true` (**owner only** in the UI):

```bash
curl -X PATCH http://localhost:8090/projects/$PID/workitems/$ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"state": "NotDone", "override": true}'
```

**Response (200):** Full updated WorkItem object.

#### GET /projects/{project_id}/workitems

List work items with filtering and pagination.

| Parameter | Type | Description |
|-----------|------|-------------|
| `state` | string | `NotDone`, `InProgress`, or `Completed` |
| `tags` | string | Comma-separated; AND logic (all must match) |
| `type` | string | Filter by type value |
| `parent_id` | UUID | Filter by parent |
| `assignee_id` | UUID | Filter by assignee (or `null` for unassigned) |
| `relationship_type` | string | `blocks`, `depends_on`, `relates_to`, `duplicate_of` |
| `relationship_target_id` | UUID | Combined with `relationship_type` |
| `is_blocked` | bool | Has unresolved `depends_on` |
| `is_ready` | bool | All `depends_on` targets are `Completed` |
| `page` | int | Default: 1 |
| `page_size` | int | Default: 50, max: 200 |

**Response (200):**

```json
{
  "items": [...],
  "total": 42,
  "page": 1,
  "page_size": 20
}
```

#### DELETE /projects/{project_id}/workitems/{id}

Deletes a work item atomically. Cascades: removes all relationships (both directions), nulls `parent_id` on children, removes tags.

**Response:** 204 No Content.

#### POST /projects/{project_id}/workitems/{id}/relationships

```bash
curl -X POST http://localhost:8090/projects/$PID/workitems/$ID/relationships \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type": "depends_on", "target_id": "660e8400-..."}'
```

**Response (201):** `{ "id": "...", "type": "depends_on", "target_id": "..." }`

#### DELETE /projects/{project_id}/workitems/{id}/relationships/{rel_id}

**Response:** 204 No Content.

#### GET /projects/{project_id}/workitems/{id}/cycles

Detects dependency cycles (`depends_on` and `blocks` edges) involving the specified work item.

**Response (200):** `{ "cycles": [["id1", "id2", ...]] }` — empty array if no cycles.

## Data Model

### User

```json
{
  "id": "UUID v4",
  "email": "string (unique, max 320 chars)",
  "display_name": "string (1-100 chars)",
  "created_at": "2026-04-07T12:00:00Z",
  "updated_at": "2026-04-07T12:00:00Z"
}
```

Password hash is never exposed in API responses.

### Project

```json
{
  "id": "UUID v4",
  "name": "string (required)",
  "description": "string",
  "created_at": "2026-04-07T12:00:00Z",
  "updated_at": "2026-04-07T12:00:00Z"
}
```

### WorkItem

```json
{
  "id": "UUID v4 (system-generated)",
  "project_id": "UUID v4",
  "title": "string (required, max 500 chars)",
  "description": "string (max 10000 chars)",
  "state": "NotDone | InProgress | Completed",
  "tags": ["string (max 100 chars each, no commas, max 50)"],
  "type": "string (optional, max 100 chars)",
  "parent_id": "UUID v4 or null",
  "assignee_id": "UUID v4 or null",
  "created_by": "UUID v4 or null",
  "assignee_name": "string (resolved from users table, read-only)",
  "relationships": [
    {
      "id": "UUID v4",
      "type": "blocks | depends_on | relates_to | duplicate_of",
      "target_id": "UUID v4"
    }
  ],
  "is_blocked": "boolean (derived)",
  "created_at": "2026-04-07T12:00:00Z (immutable)",
  "updated_at": "2026-04-07T12:00:00Z (auto-updated)"
}
```

### Project Roles

| Role | Permissions |
|------|-------------|
| `owner` | Full access: manage members, edit/delete project, all work item operations, override state transitions |
| `member` | Create, edit, and delete work items; manage relationships and tags |
| `viewer` | Read-only access to work items, relationships, and project data |

### State Machine

```
NotDone ──> InProgress ──> Completed
   ^             ^              │
   └─────────────┴── override ──┘
```

- Forward (one step): always allowed
- Backward or skip: requires `override: true` in the request body

### Database Schema

Six migrations create the following tables:

- **work_items** — Core work item data with indexes on state, type, parent_id, assignee_id, created_by
- **work_item_tags** — Composite PK (item_id, tag), FK to work_items with CASCADE
- **work_item_relationships** — Unique constraint on (source_id, target_id, type), FKs with CASCADE
- **projects** — Project metadata, FK from work_items.project_id
- **users** — User accounts with unique email constraint, bcrypt password hash
- **project_memberships** — Composite unique on (project_id, user_id), role column

Migrations are applied automatically on startup using advisory locking to prevent concurrent execution.

## Architecture Notes

- **Store layer** uses batch loading (3 queries total for List, not N+1) and recursive CTEs for hierarchy traversal
- **Cycle detection** loads the reachable subgraph via recursive CTE in a single query, then runs DFS in memory
- **Migration runner** uses MySQL advisory locks (`GET_LOCK`/`RELEASE_LOCK`) on a dedicated connection to prevent concurrent migrations
- **No transactions for DDL** — MySQL implicitly commits DDL statements, so each migration file contains a single DDL operation
- **Authentication** uses HS256 JWT with short-lived access tokens (15m default) and long-lived refresh tokens (7d) stored in HttpOnly cookies
- **Authorization** is enforced at two levels: middleware extracts the user's project role, and handler helpers (`requireOwner`, `requireWriteAccess`) gate specific operations
- **Frontend auth** stores the access token in memory (not localStorage) and uses a shared-promise pattern to deduplicate concurrent token refresh requests
