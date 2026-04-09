# Multi-Project Support — Specification

## 1. Overview

Lattice currently operates as a single flat workspace: all work items share one global pool. This spec adds **project** scoping so that work items, relationships, and tags are partitioned by project. Users can create multiple projects and switch between them.

## 2. Goals

- Work items belong to exactly one project.
- API routes are scoped under `/projects/{project_id}/workitems/...`.
- A top-level `/projects` CRUD API manages projects.
- Relationships and cycle detection are project-scoped (no cross-project references).
- The frontend supports project switching with a project selector.
- Backward compatible: a migration assigns all existing work items to a default project.

## 3. Non-Goals

- Cross-project relationships or dependencies.
- Moving work items between projects.
- Project-level permissions or access control.
- Project archival or soft-delete (may be added later).
- Nested projects / sub-projects.

## 4. Domain Model

### 4.1 Project Entity

| Field | Type | Constraints |
|-------|------|-------------|
| `id` | CHAR(36) | UUID v4, primary key, system-generated |
| `name` | VARCHAR(200) | Required, 1–200 characters, unique |
| `description` | TEXT | Optional (defaults to empty string), max 5,000 characters |
| `created_at` | DATETIME(3) | System-generated, immutable |
| `updated_at` | DATETIME(3) | System-managed |

### 4.2 WorkItem Changes

- Add `project_id CHAR(36) NOT NULL` column to `work_items`, with a foreign key to `projects(id)`.
- Add index on `project_id`.
- Parent-child hierarchy is project-scoped: `parent_id` must reference a work item in the same project.
- Tags remain per-item (no change to `work_item_tags`).

### 4.3 Relationship Changes

- Relationships are implicitly project-scoped because both `source_id` and `target_id` must be work items in the same project.
- The `AddRelationship` handler must validate that `target_id` belongs to the same project as the source work item.

## 5. API

### 5.1 Project Endpoints

All project endpoints live at the top level.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/projects` | Create a project |
| GET | `/projects` | List all projects |
| GET | `/projects/{project_id}` | Get a project |
| PATCH | `/projects/{project_id}` | Update a project |
| DELETE | `/projects/{project_id}` | Delete a project (must have zero work items) |

#### POST /projects

**Request:**
```json
{
  "name": "Backend Services",
  "description": "All backend microservice work"
}
```

**Response (201):** Full project object with generated `id`, `created_at`, `updated_at`.

#### GET /projects

**Response (200):**
```json
{
  "projects": [
    { "id": "...", "name": "...", "description": "...", "item_count": 12, "created_at": "...", "updated_at": "..." }
  ]
}
```

The `item_count` field is the total count of all work items in the project (regardless of hierarchy level).

#### PATCH /projects/{project_id}

**Request:** Partial update — `name` and/or `description`. Only provided fields are modified.

**Response (200):** Full project object.

#### DELETE /projects/{project_id}

Fails with 409 CONFLICT if the project contains any work items. The client must delete all items first.

**Note:** Work items cannot be moved between projects. Moving would violate project-scoped parent and relationship invariants. To relocate work, delete and recreate in the target project.

**Response:** 204 No Content.

### 5.2 Work Item Endpoints (Re-scoped)

All existing work item endpoints move under the project scope:

| Method | Path |
|--------|------|
| POST | `/projects/{project_id}/workitems` |
| GET | `/projects/{project_id}/workitems` |
| GET | `/projects/{project_id}/workitems/{id}` |
| PATCH | `/projects/{project_id}/workitems/{id}` |
| DELETE | `/projects/{project_id}/workitems/{id}` |
| POST | `/projects/{project_id}/workitems/{id}/relationships` |
| DELETE | `/projects/{project_id}/workitems/{id}/relationships/{rel_id}` |
| GET | `/projects/{project_id}/workitems/{id}/cycles` |

**Behavior changes:**
- Create sets `project_id` from the URL path (not the request body).
- Get/Update/Delete verify the work item belongs to the project in the URL (404 if mismatch).
- List filters only work items in the specified project.
- AddRelationship validates `target_id` belongs to the same project.
- Parent assignment validates the parent belongs to the same project.

### 5.3 Error Codes

| HTTP Status | Code | New Cause |
|-------------|------|-----------|
| 404 | `NOT_FOUND` | Project not found |
| 409 | `CONFLICT` | Delete project with existing work items, or duplicate project name |
| 422 | `VALIDATION_ERROR` | Cross-project parent or relationship target |

## 6. Database Migration

### 6.1 Migration: Create projects table

```sql
CREATE TABLE projects (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uq_projects_name (name)
);
```

### 6.2 Migration: Add project_id to work_items

```sql
-- Create default project for existing data
INSERT INTO projects (id, name, description)
VALUES ('00000000-0000-0000-0000-000000000001', 'Default', 'Auto-created project for existing work items');

-- Add column, backfill, add constraint
ALTER TABLE work_items ADD COLUMN project_id CHAR(36) NULL AFTER id;
UPDATE work_items SET project_id = '00000000-0000-0000-0000-000000000001' WHERE project_id IS NULL;
ALTER TABLE work_items MODIFY COLUMN project_id CHAR(36) NOT NULL;
ALTER TABLE work_items ADD INDEX idx_work_items_project_id (project_id);
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_project FOREIGN KEY (project_id) REFERENCES projects(id);
```

### 6.3 Down Migration

```sql
ALTER TABLE work_items DROP FOREIGN KEY fk_work_items_project;
ALTER TABLE work_items DROP INDEX idx_work_items_project_id;
ALTER TABLE work_items DROP COLUMN project_id;
DROP TABLE projects;
```

## 7. Frontend

### 7.1 Project Selector

- A project selector in the app shell header/sidebar allows switching the active project.
- The selected project ID is stored in the URL path (e.g., `/projects/{id}/board`).
- All API calls use the project ID from the current route.

### 7.2 Project Management Page

- A `/projects` page lists all projects with name, description, and item count.
- Create/edit project via modal form.
- Delete project (only if empty) with confirmation.

### 7.3 Route Changes

| Old Route | New Route |
|-----------|-----------|
| `/` | `/` (project list / home) |
| `/board` | `/projects/{id}/board` |
| `/list` | `/projects/{id}/list` |
| `/graph` | `/projects/{id}/graph` |
| `/items/{id}` | `/projects/{id}/workitems/{itemId}` |
| `/settings` | `/settings` (global, unchanged) |

### 7.4 Type Changes

- Add `Project` type: `{ id, name, description, item_count, created_at, updated_at }`.
- Add `project_id` to `WorkItem` type.
- API client functions accept `projectId` as first parameter.

## 8. Invariants

1. Every work item belongs to exactly one project.
2. Parent-child relationships are project-scoped: a work item's parent must be in the same project.
3. Dependency/block relationships are project-scoped: both source and target must be in the same project.
4. A project cannot be deleted while it contains work items.
5. Project names are unique.
6. Cycle detection operates within project scope only.
7. The auto-created "Default" project (ID `00000000-0000-0000-0000-000000000001`) follows the same rules as any other project — it can be renamed or deleted (if empty).
