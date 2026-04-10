# Multi-User Support — Specification

## 1. Overview

Lattice currently operates as a single-user system. An `X-Role` header distinguishes admin from regular users, but there is no concept of user identity, authentication, or per-user authorization. This specification adds multi-user support: user accounts, authentication via JWT, project-level role-based access control, and work item assignment.

## 2. Goals

1. Users can register and authenticate with email/password.
2. Every API request is associated with an authenticated user (except public endpoints).
3. Projects have membership with roles: **owner**, **member**, **viewer**.
4. Work items can be assigned to a user.
5. The existing `X-Role: admin` mechanism is replaced by per-project roles.
6. The system remains simple — no organizations, teams, or nested permission hierarchies.

## 3. Non-Goals

- OAuth / SSO / third-party identity providers (future work).
- Fine-grained per-field permissions.
- Audit log of user actions (future work).
- User profile avatars or rich profile data.
- Email verification or password reset flows (future work).

## 4. Domain Model

### 4.1 User

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK, generated |
| email | string | unique, max 320 chars, valid email format |
| display_name | string | 1–100 chars |
| password_hash | string | bcrypt hash, never exposed via API |
| created_at | datetime | auto |
| updated_at | datetime | auto |

### 4.2 ProjectMembership

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK, generated |
| project_id | UUID | FK → projects.id |
| user_id | UUID | FK → users.id |
| role | enum | `owner`, `member`, `viewer` |
| created_at | datetime | auto |

- Composite unique constraint on `(project_id, user_id)`.
- Every project must have exactly one owner. The creator becomes the owner.
- The owner can transfer ownership but cannot remove themselves without transferring first.

### 4.3 WorkItem Changes

Add an optional `assignee_id` field:

| Field | Type | Constraints |
|-------|------|-------------|
| assignee_id | UUID \| null | FK → users.id, nullable |

### 4.4 Role Permissions

| Action | owner | member | viewer |
|--------|-------|--------|--------|
| View project and work items | yes | yes | yes |
| Create work items | yes | yes | no |
| Update work items (forward transitions) | yes | yes | no |
| Delete work items | yes | yes | no |
| Override state transitions (backward) | yes | no | no |
| Manage relationships | yes | yes | no |
| Update project settings | yes | no | no |
| Delete project | yes | no | no |
| Invite/remove members | yes | no | no |
| Change member roles | yes | no | no |

## 5. Authentication

### 5.1 Mechanism

- **JWT bearer tokens** in the `Authorization: Bearer <token>` header.
- Access tokens expire after **15 minutes**.
- Refresh tokens expire after **7 days**, stored as HTTP-only cookies.
- Tokens contain: `sub` (user ID), `exp`, `iat`.
- Token signing uses HMAC-SHA256 with a server-side secret (`LATTICE_JWT_SECRET` env var).

### 5.2 Public Endpoints (No Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| POST | /auth/register | Create account |
| POST | /auth/login | Authenticate, receive tokens |
| POST | /auth/refresh | Exchange refresh token for new access token |

### 5.3 Protected Endpoints

All existing endpoints (`/projects/...`) require a valid access token. The middleware extracts the user ID from the token and attaches it to the request context.

### 5.4 Registration

```
POST /auth/register
{
  "email": "user@example.com",
  "display_name": "Alice",
  "password": "..."
}
```

- Password minimum 8 characters.
- Returns the created user (without password_hash) and tokens.
- If email already exists → `409 CONFLICT`.

### 5.5 Login

```
POST /auth/login
{
  "email": "user@example.com",
  "password": "..."
}
```

- Returns access token in body + refresh token as HTTP-only cookie.
- Invalid credentials → `401 UNAUTHORIZED`.

### 5.6 Refresh

```
POST /auth/refresh
Cookie: lattice_refresh=<token>
```

- Returns new access token.
- If refresh token is expired or invalid → `401 UNAUTHORIZED`.

## 6. Authorization

### 6.1 Middleware Chain

The middleware chain becomes:

```
Logging → Auth → ProjectRole → ContentType → Handler
```

1. **Auth middleware**: Validates JWT, extracts user ID, attaches to context. Skips public endpoints.
2. **ProjectRole middleware**: For project-scoped routes, loads the user's role for that project and attaches to context. Returns `403 FORBIDDEN` if the user has no membership.

### 6.2 Override Behavior Change

The current `X-Role: admin` header and `override` field are replaced:

- Only project **owners** may use `override: true` for backward state transitions.
- The `X-Role` header is **removed** from the system.
- The `RoleMiddleware` is replaced by `AuthMiddleware` + `ProjectRoleMiddleware`.

### 6.3 Project Creation

Any authenticated user can create a project. The creator is automatically added as the owner.

### 6.4 Default Project Migration

The default project (ID `00000000-0000-0000-0000-000000000001`) created during multi-project migration needs an owner. The migration creates a system user (or the first registered user becomes the owner). For existing deployments, the migration adds all existing users as owners of the default project.

## 7. API Changes

### 7.1 New Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /auth/register | Register |
| POST | /auth/login | Login |
| POST | /auth/refresh | Refresh token |
| GET | /users/me | Get current user |
| PATCH | /users/me | Update display_name or password |
| GET | /projects/{id}/members | List project members |
| POST | /projects/{id}/members | Add member (owner only) |
| PATCH | /projects/{id}/members/{user_id} | Change role (owner only) |
| DELETE | /projects/{id}/members/{user_id} | Remove member (owner only) |

### 7.2 Modified Endpoints

- `POST /projects` — Creator becomes owner automatically.
- `PATCH /projects/{id}/workitems/{id}` — `override` requires owner role (not admin header).
- `GET /projects` — Returns only projects the user is a member of.

### 7.3 Work Item Assignment

- `POST /projects/{id}/workitems` — Optional `assignee_id` field.
- `PATCH /projects/{id}/workitems/{id}` — Can set/clear `assignee_id`.
- `GET /projects/{id}/workitems?assignee_id=<uuid>` — Filter by assignee.
- Assignee must be a member of the project (member or owner, not viewer).

### 7.4 Response Changes

- All user-facing responses include `created_by` user ID on work items.
- Project list response includes the current user's role.
- Work item responses include `assignee` object (id + display_name) when assigned.

## 8. Frontend Changes

### 8.1 Auth Pages

- Login page at `/login`.
- Register page at `/register`.
- Unauthenticated users are redirected to `/login`.

### 8.2 Auth State

- Access token stored in memory (not localStorage).
- Refresh token stored as HTTP-only cookie (managed by browser).
- Auth context provides: `user`, `isAuthenticated`, `login()`, `logout()`, `register()`.

### 8.3 Project Members UI

- Members panel on project settings page.
- Invite form: email input + role selector.
- Member list: shows user, role, remove button (owner only).

### 8.4 Work Item Assignment

- Assignee dropdown on work item detail page.
- Filter by "Assigned to me" on list/board pages.
- Avatar/initials badge on work item cards.

### 8.5 Role-Based UI

- Hide create/edit controls for viewers.
- Hide delete/settings for non-owners.
- Hide override toggle for non-owners.
- Show user's role badge in project header.

## 9. Database Schema

### 9.1 New Tables

```sql
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(72) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uq_users_email (email)
);

CREATE TABLE project_memberships (
    id CHAR(36) PRIMARY KEY,
    project_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    role ENUM('owner', 'member', 'viewer') NOT NULL DEFAULT 'member',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_membership (project_id, user_id)
);
```

### 9.2 Work Items Modification

```sql
ALTER TABLE work_items ADD COLUMN assignee_id CHAR(36) NULL;
ALTER TABLE work_items ADD COLUMN created_by CHAR(36) NULL;
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE work_items ADD INDEX idx_work_items_assignee (assignee_id);
```

## 10. Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| LATTICE_JWT_SECRET | yes | — | HMAC-SHA256 signing key (min 32 chars) |
| LATTICE_ACCESS_TOKEN_TTL | no | 15m | Access token lifetime |
| LATTICE_REFRESH_TOKEN_TTL | no | 168h | Refresh token lifetime |

## 11. Invariants

1. Every project has exactly one owner at all times.
2. A user cannot be a member of the same project twice (unique constraint).
3. Only project owners can invite, remove, or change roles of members.
4. Only project owners can use `override: true` for backward state transitions.
5. Viewers cannot create, update, or delete work items or relationships.
6. Assignees must be members of the project (owner or member role).
7. Deleting a user sets `assignee_id` and `created_by` to NULL on their work items (ON DELETE SET NULL).
8. Access tokens are stateless (validated by signature + expiry only).
9. The `X-Role` header is no longer used after migration; all authorization is derived from project membership roles.
10. Passwords are never stored in plaintext or returned in API responses.
