# Multi-Project Support — Implementation Plan

**Spec:** `specs/multi-project/SPEC.md`

## Phase 1 — Domain Model and Database Migration

### Goal
Add the `Project` entity to the domain layer and create database migrations.

### Tasks

1. **Domain model** — Create `internal/domain/project.go`:
   - `Project` struct: ID, Name, Description, CreatedAt, UpdatedAt
   - `Validate()` method: name 1–200 chars, description max 5,000 chars
   - Add `ProjectID` field to `WorkItem` struct in `workitem.go`

2. **Store interface** — Add `ProjectStore` to `internal/store/store.go`:
   - `Create(ctx, project) error`
   - `Get(ctx, id) (*Project, error)`
   - `Update(ctx, id, params) (*Project, error)`
   - `Delete(ctx, id) error`
   - `List(ctx) ([]ProjectWithCount, error)` — includes item_count

3. **Migration 005** — `migrations/005_create_projects.up.sql`:
   - CREATE TABLE projects (id, name, description, created_at, updated_at, unique name)
   - Insert default project with well-known ID
   - ALTER work_items: add project_id, backfill, add NOT NULL + FK + index
   - Down migration: drop FK, drop column, drop table

4. **Add `ErrConflict`** sentinel error to `internal/domain/errors.go`

### Verification
- `go build ./...` passes
- Migration applies and rolls back cleanly on test database
- Existing tests still pass (WorkItem struct gains ProjectID field)

---

## Phase 2 — Project MySQL Store

### Goal
Implement the MySQL-backed project store.

### Tasks

1. **Create `internal/store/mysql/project.go`**:
   - `ProjectStore` struct with `*sql.DB`
   - `Create`: generate UUID, validate, INSERT
   - `Get`: SELECT by ID
   - `Update`: partial update (name, description), handle duplicate name → ErrConflict
   - `Delete`: attempt DELETE and rely on FK constraint; catch MySQL error 1451 → ErrConflict
   - `List`: SELECT with COUNT subquery for item_count, ordered by name

2. **Handle MySQL errors**:
   - Error 1062 (duplicate key on name) → `domain.ErrConflict`
   - Error 1451 (FK constraint on delete) → `domain.ErrConflict`

### Verification
- Unit tests for all CRUD operations
- Duplicate name test
- Delete-with-items rejection test

---

## Phase 3 — Scope Work Item Store to Projects

### Goal
Add project_id filtering to all work item operations.

### Tasks

1. **Update `WorkItemStore` interface**:
   - `Create`: set project_id from parameter
   - `Get`: add projectID parameter, verify work item belongs to project
   - `Update`: add projectID parameter, verify ownership
   - `Delete`: add projectID parameter, verify ownership
   - `List`: add projectID to `ListFilter`, always filter by project

2. **Update `internal/store/mysql/workitem.go`**:
   - `Create`: INSERT includes project_id
   - `Get`: WHERE clause includes `project_id = ?`
   - `Update`: WHERE clause includes `project_id = ?`
   - `Delete`: WHERE clause includes `project_id = ?`
   - `List`: WHERE clause always includes `project_id = ?`
   - Parent assignment: validate parent's project_id matches

3. **Update `RelationshipStore.Add`**:
   - Before inserting, SELECT target work item's project_id
   - If target's project_id ≠ source's project_id → `ErrValidation`

4. **Update `CycleDetector`**:
   - Scope recursive CTE to project_id (add JOIN or WHERE)

5. **Update existing tests** to pass project_id

### Verification
- All existing tests pass with project scoping
- Cross-project parent assignment rejected
- Cross-project relationship rejected

---

## Phase 4 — Project API Handlers and Re-scoped Routes

### Goal
Add project CRUD endpoints and move work item routes under `/projects/{project_id}/`.

### Tasks

1. **Add project handlers** to `internal/api/handler.go`:
   - `CreateProject` — POST `/projects`
   - `ListProjects` — GET `/projects`
   - `GetProject` — GET `/projects/{project_id}`
   - `UpdateProject` — PATCH `/projects/{project_id}`
   - `DeleteProject` — DELETE `/projects/{project_id}`

2. **Re-scope work item routes**:
   - Change all routes from `/workitems/...` to `/projects/{project_id}/workitems/...`
   - Extract `project_id` from URL path in each handler
   - Pass project_id to store methods

3. **Update error mapping** in `errors.go`:
   - Map `ErrConflict` → 409 CONFLICT
   - Verify `ErrValidation` → 422 VALIDATION_ERROR is already mapped (it should be via existing ErrValidation handling)

4. **Add `ProjectStore` field** to `Handler` struct

5. **Update `cmd/lattice/main.go`**:
   - Create `ProjectStore` and pass to Handler

### Verification
- `curl` tests for all project CRUD endpoints
- Existing work item API tests updated for new routes
- 409 on duplicate project name
- 409 on delete project with items
- 404 on work item in wrong project

---

## Phase 5 — Frontend API Client and Types

### Goal
Update the frontend API client layer for project-scoped endpoints.

### Tasks

1. **Update `frontend/src/lib/types.ts`**:
   - Add `Project` type
   - Add `project_id` to `WorkItem` type

2. **Create `frontend/src/lib/api/projects.ts`**:
   - `listProjects(): Promise<ProjectListResponse>`
   - `getProject(id): Promise<Project>`
   - `createProject(input): Promise<Project>`
   - `updateProject(id, input): Promise<Project>`
   - `deleteProject(id): Promise<void>`

3. **Update `frontend/src/lib/api/workitems.ts`**:
   - All functions take `projectId` as first parameter
   - URL paths become `/projects/${projectId}/workitems/...`

4. **Update `frontend/src/lib/api/relationships.ts`** and **`cycles.ts`**:
   - Add `projectId` parameter, update URL paths

5. **Add `frontend/src/hooks/useProjects.ts`**:
   - `useProjects()` — list query
   - `useProject(id)` — single query
   - `useProjectMutations()` — create, update, delete

6. **Update `frontend/src/hooks/useWorkItems.ts`**:
   - Accept `projectId`, pass to API calls

7. **Update validation schemas** if needed

### Verification
- TypeScript compiles cleanly
- Existing hook tests updated

---

## Phase 6 — Frontend Routing and Project Selector

### Goal
Add project-scoped routing and a project selector to the UI.

### Tasks

1. **Update router** (`frontend/src/app/router.tsx`):
   - `/` — project list (home)
   - `/projects/:projectId/board` — board view
   - `/projects/:projectId/list` — list view
   - `/projects/:projectId/graph` — graph view
   - `/projects/:projectId/workitems/:itemId` — item detail
   - `/settings` — unchanged

2. **Create `useProjectId()` hook**:
   - Extract `projectId` from `useParams()`
   - Used by all project-scoped pages

3. **Update AppShell**:
   - Add project selector dropdown in header
   - Show current project name
   - Navigates to same view in selected project on switch

4. **Update all pages** (BoardPage, ListPage, GraphPage, ItemDetailPage, HomePage):
   - Use `useProjectId()` to get current project
   - Pass to hooks and API calls

5. **Create ProjectsPage**:
   - List all projects with item counts
   - Create/edit/delete project modals
   - Click project → navigate to its board

### Verification
- Navigation between projects works
- URL reflects current project
- All pages load correctly with project scope
- Creating/editing/deleting projects works
