# Multi-User Support — Implementation Plan

## Prerequisites

- Multi-project support (Phases 1–6) is complete and merged.
- Go dependencies: `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`.

---

## Phase 1 — User Domain Model and Store

### Goal
Define the User entity, store interface, and MySQL implementation with migration.

### Tasks

1. **Create `internal/domain/user.go`**:
   - `User` struct: ID, Email, DisplayName, PasswordHash, CreatedAt, UpdatedAt.
   - `Validate()` — email format, display_name length (1–100), email length (≤320).
   - Password validation helper: `ValidatePassword(plain string) error` — min 8 chars.
   - Do NOT expose PasswordHash in JSON (`json:"-"` tag).

2. **Create `internal/domain/membership.go`**:
   - `ProjectRole` type: `owner`, `member`, `viewer`.
   - `ValidProjectRole(r ProjectRole) bool`.
   - `ProjectMembership` struct: ID, ProjectID, UserID, Role, CreatedAt.

3. **Update `internal/domain/errors.go`**:
   - Add `ErrUnauthorized = errors.New("unauthorized")`.
   - Add `ErrDuplicateEmail = errors.New("duplicate email")` (wraps ErrConflict).

4. **Add store interfaces to `internal/store/store.go`**:
   - `UserStore`: Create, GetByID, GetByEmail, UpdateDisplayName, UpdatePassword, Delete.
   - `MembershipStore`: Add, Remove, UpdateRole, ListByProject, GetRole(ctx, projectID, userID).

5. **Create `internal/store/mysql/user.go`**:
   - Implement UserStore. Hash password with bcrypt (cost 12) in Create.
   - Map duplicate email MySQL error (1062 on uq_users_email) to ErrDuplicateEmail.

6. **Create `internal/store/mysql/membership.go`**:
   - Implement MembershipStore.
   - `GetRole` returns the user's role for a project, or ErrNotFound if not a member.
   - `Remove` prevents removing the last owner.

7. **Create migration `006_add_users.up.sql`** and corresponding down migration:
   - `users` table with email unique constraint.
   - `project_memberships` table with composite unique on (project_id, user_id).
   - Add `assignee_id` and `created_by` columns to `work_items`.
   - Collation: `utf8mb4_unicode_ci` to match existing tables.

8. **Unit tests** for User.Validate(), ProjectRole validation, store implementations.

### Verification
- `go build ./... && go test ./... && golangci-lint run ./...`
- Migration applies and rolls back cleanly.

---

## Phase 2 — Authentication (JWT)

### Goal
Implement registration, login, token issuance, and auth middleware.

### Tasks

1. **Create `internal/auth/token.go`**:
   - `TokenService` struct: secret, accessTTL, refreshTTL.
   - `NewTokenService(secret string, accessTTL, refreshTTL time.Duration)`.
   - `IssueAccessToken(userID string) (string, error)` — JWT with `sub`, `exp`, `iat`.
   - `IssueRefreshToken(userID string) (string, error)` — longer-lived JWT.
   - `ValidateToken(tokenStr string) (userID string, err error)`.
   - Use `github.com/golang-jwt/jwt/v5` with HS256.

2. **Create `internal/auth/password.go`**:
   - `HashPassword(plain string) (string, error)` — bcrypt cost 12.
   - `CheckPassword(hash, plain string) error`.

3. **Create `internal/api/auth_handler.go`**:
   - `AuthHandler` struct: UserStore, TokenService.
   - `Register(w, r)` — parse body, validate, create user, issue tokens, return user + access token, set refresh cookie.
   - `Login(w, r)` — verify email/password, issue tokens.
   - `Refresh(w, r)` — read refresh cookie, validate, issue new access token.
   - Register routes: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`.

4. **Create `internal/api/auth_middleware.go`**:
   - `AuthMiddleware(tokenService, next)` — extract Bearer token, validate, set userID in context.
   - Skip auth for `/auth/` prefix paths.
   - `UserIDFromContext(ctx) string` helper.
   - Return `401 UNAUTHORIZED` for missing/invalid tokens.

5. **Create `internal/api/user_handler.go`**:
   - `GET /users/me` — return current user from context.
   - `PATCH /users/me` — update display_name or password.

6. **Update `cmd/lattice/main.go`**:
   - Read `LATTICE_JWT_SECRET` env var (required).
   - Create TokenService, AuthHandler.
   - Insert AuthMiddleware into chain: `Logging → Auth → JSONContentType → mux`.
   - Remove old `RoleMiddleware` from chain.

7. **Unit tests** for token issuance/validation, password hashing, auth handlers with mock stores.

### Verification
- `go build ./... && go test ./... && golangci-lint run ./...`
- Can register, login, and access protected endpoints with Bearer token.

---

## Phase 3 — Project Role Authorization

### Goal
Enforce project-level permissions based on membership roles.

### Tasks

1. **Create `internal/api/project_role_middleware.go`**:
   - For routes matching `/projects/{project_id}/...`, load the user's role via MembershipStore.GetRole().
   - Attach role to context: `ProjectRoleFromContext(ctx) ProjectRole`.
   - Return `403 FORBIDDEN` if user has no membership.
   - Skip for non-project routes (`/auth/`, `/users/me`, `GET /projects` list).

2. **Create `internal/api/authz.go`** — authorization helpers:
   - `requireRole(ctx, minRole ProjectRole) error` — checks context role against minimum.
   - `requireWriteAccess(ctx) error` — requires member or owner.
   - `requireOwner(ctx) error` — requires owner.

3. **Update work item handlers in `handler.go`**:
   - `CreateWorkItem` — call `requireWriteAccess`.
   - `UpdateWorkItem` — call `requireWriteAccess`. If `override: true`, call `requireOwner`.
   - `DeleteWorkItem` — call `requireWriteAccess`.
   - `AddRelationship` / `RemoveRelationship` — call `requireWriteAccess`.

4. **Update project handlers**:
   - `UpdateProject` — call `requireOwner`.
   - `DeleteProject` — call `requireOwner`.
   - `ListProjects` — filter to only projects the user is a member of.
   - `CreateProject` — auto-add creator as owner via MembershipStore.

5. **Add member management handlers**:
   - `GET /projects/{id}/members` — list members (any role can view).
   - `POST /projects/{id}/members` — add member (owner only). Body: `{ "email": "...", "role": "member" }`.
   - `PATCH /projects/{id}/members/{user_id}` — change role (owner only).
   - `DELETE /projects/{id}/members/{user_id}` — remove member (owner only). Prevent removing last owner.

6. **Remove old RoleMiddleware** (`middleware.go`):
   - Delete `RoleMiddleware` function and `roleKey` context key.
   - Remove `X-Role` header handling from middleware chain.

7. **Update `internal/api/errors.go`**:
   - Add `ErrUnauthorized` → 401 mapping.

8. **Update handler_test.go and integration_test.go**:
   - Add auth context to test requests.
   - Test permission enforcement: viewer cannot create, member cannot override, owner can do all.

### Verification
- `go build ./... && go test ./... && golangci-lint run ./...`
- Viewers get 403 on write attempts. Members get 403 on override. Owners succeed.

---

## Phase 4 — Work Item Assignment

### Goal
Add assignee support to work items.

### Tasks

1. **Update `internal/domain/workitem.go`**:
   - Add `AssigneeID *string` and `CreatedBy *string` fields.
   - Add `AssigneeName string` (populated by store, not persisted — `json:"assignee_name,omitempty"`).

2. **Update `internal/store/store.go`**:
   - Add `AssigneeID *string` to `UpdateParams`.
   - Add `AssigneeID *string` to `ListFilter`.
   - Add `CreatedBy` to `CreateParams` or set in store.

3. **Update `internal/store/mysql/workitem.go`**:
   - Include `assignee_id`, `created_by` in all queries.
   - JOIN users table to get assignee display_name on Get/List.
   - On Create: set `created_by` from context (passed via work item struct).
   - On Update with assignee_id: validate assignee is a project member (owner or member).

4. **Update API handlers**:
   - `CreateWorkItem` — set `CreatedBy` from auth context.
   - `UpdateWorkItem` — accept `assignee_id` field.
   - `ListWorkItems` — support `?assignee_id=<uuid>` filter.

5. **Unit tests** for assignment validation, filtering.

### Verification
- `go build ./... && go test ./... && golangci-lint run ./...`
- Can assign, reassign, and filter by assignee.

---

## Phase 5 — Frontend Authentication

### Goal
Add login/register pages and auth state management.

### Tasks

1. **Create `frontend/src/lib/api/auth.ts`**:
   - `register(email, displayName, password)` → returns user + access token.
   - `login(email, password)` → returns user + access token.
   - `refresh()` → returns new access token.
   - `getMe()` → returns current user.
   - `updateMe(input)` → update profile.

2. **Create `frontend/src/hooks/useAuth.ts`** (React context):
   - `AuthProvider` — wraps app, manages token state.
   - Stores access token in ref (not state, to avoid re-renders).
   - On mount: attempt `refresh()` to restore session.
   - `login()`, `logout()`, `register()` methods.
   - `user`, `isAuthenticated` state.

3. **Update `frontend/src/lib/api/client.ts`**:
   - `apiFetch` reads token from auth context/module.
   - Sets `Authorization: Bearer <token>` header.
   - On 401 response: attempt refresh, retry once, else redirect to login.
   - Remove `X-Role` header logic and `getRole()` dependency.

4. **Create `frontend/src/pages/LoginPage.tsx`**:
   - Email + password form.
   - Error display for invalid credentials.
   - Link to register page.

5. **Create `frontend/src/pages/RegisterPage.tsx`**:
   - Email + display name + password + confirm password form.
   - Validation feedback.
   - Link to login page.

6. **Update `frontend/src/app/router.tsx`**:
   - Add `/login` and `/register` routes (public).
   - Add `ProtectedRoute` wrapper that redirects to `/login` if not authenticated.
   - Wrap project routes in `ProtectedRoute`.

7. **Remove `frontend/src/lib/config.ts`** role management:
   - Delete `getRole()`, `isAdmin()`, `window.__LATTICE_CONFIG__`.
   - Role is now derived from project membership.

8. **Update tests** to mock auth context.

### Verification
- TypeScript compiles cleanly.
- Can register, login, and access project pages.
- Unauthenticated access redirects to login.

---

## Phase 6 — Frontend Authorization and Member Management

### Goal
Add project member management UI and role-based UI controls.

### Tasks

1. **Create `frontend/src/hooks/useProjectRole.ts`**:
   - Fetches current user's role for the active project.
   - `useProjectRole()` → `{ role, isOwner, canWrite }`.
   - Derive from project members list or dedicated endpoint.

2. **Create `frontend/src/lib/api/members.ts`**:
   - `listMembers(projectId)`, `addMember(projectId, email, role)`.
   - `updateMemberRole(projectId, userId, role)`, `removeMember(projectId, userId)`.

3. **Create `frontend/src/hooks/useMembers.ts`**:
   - `useMembers(projectId)` — list query.
   - `useMemberMutations(projectId)` — add, update role, remove.

4. **Create `frontend/src/pages/MembersPage.tsx`** (or section in project settings):
   - Member list with role badges.
   - Invite form (email + role dropdown) — owner only.
   - Change role / remove buttons — owner only.

5. **Update existing pages for role-based visibility**:
   - `BoardPage` — hide drag-and-drop for viewers.
   - `ListPage` — hide action buttons for viewers.
   - `ItemDetailPage` — hide edit controls for viewers, hide override for non-owners.
   - `CreateWorkItemForm` — hide entirely for viewers.
   - `QuickAdd` — hide for viewers.
   - `AppShell` — hide "Create" button for viewers.

6. **Add assignee UI**:
   - `AssigneeSelector` component — dropdown of project members.
   - Add to `ItemDetailPage` sidebar.
   - Add "Assigned to me" quick filter on list/board pages.
   - Show assignee initials on `WorkItemCard` and `CompactCard`.

7. **Update `ProjectsPage`** — show user's role badge per project.

8. **Update frontend types**:
   - Add `User`, `ProjectMember` types.
   - Add `assignee_id`, `created_by`, `assignee_name` to `WorkItem`.
   - Add `role` to project list response.

### Verification
- Viewers see read-only UI.
- Members can create/edit.
- Owners see all controls including member management.
- Assignee picker works and filters correctly.

---

## Phase Summary

| Phase | Scope | Key Files |
|-------|-------|-----------|
| 1 | User model, membership store, migration | domain/user.go, domain/membership.go, store/mysql/user.go, store/mysql/membership.go, migration 006 |
| 2 | JWT auth, register/login, auth middleware | auth/token.go, auth/password.go, api/auth_handler.go, api/auth_middleware.go |
| 3 | Project role enforcement, member management API | api/project_role_middleware.go, api/authz.go, handler updates |
| 4 | Work item assignment | domain/workitem.go, store/mysql/workitem.go, handler updates |
| 5 | Frontend auth (login, register, token management) | lib/api/auth.ts, hooks/useAuth.ts, LoginPage, RegisterPage |
| 6 | Frontend authorization UI, members, assignees | hooks/useProjectRole.ts, MembersPage, role-gated UI |

## Migration Strategy

- Phase 1 migration (006) adds tables and columns only — no data migration needed.
- Existing work items get `assignee_id = NULL` and `created_by = NULL` (acceptable).
- After deploying Phase 2, the first registered user should be made owner of the default project.
- The `LATTICE_JWT_SECRET` env var must be set before deploying Phase 2.
