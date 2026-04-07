# SPEC.md

# Project
Lattice Frontend

## Purpose
Build a modern web frontend for Lattice, a lightweight development work tracking system intended to replace Jira for teams that want power without bloat.

The frontend must make the system feel simple, fast, and obvious for end users, while hiding underlying model complexity such as graphs, relationships, and derived behaviors unless explicitly requested.

This specification is for the frontend application only. It assumes a backend API exists or will exist separately.

---

# Product Goals

1. Make everyday work tracking dramatically simpler than Jira
2. Preserve support for real development workflows without exposing unnecessary complexity
3. Prioritize speed, clarity, and low cognitive load
4. Make hierarchy and relationships understandable without forcing users into workflow gymnastics
5. Allow advanced structure to exist beneath a calm, minimal interface
6. Be highly usable by software teams, including engineers, tech leads, product managers, and QA
7. Be suitable for future AI-assisted features without requiring them in the first implementation

---

# Product Philosophy

The frontend must follow these principles:

## 1. Simplicity First
The primary UI must expose only the concepts most users need every day:
- Work items
- State
- Tags
- Parent-child grouping
- Basic relationships

## 2. Complexity on Demand
Advanced concepts such as dependency graphs, reverse relations, and relationship metadata must be accessible, but not pushed into the user’s face.

## 3. Fast Over Fancy
The product should feel responsive, lightweight, and professional. Avoid gratuitous animation, clutter, and dashboard theater.

## 4. Opinionated UX
Do not expose endless customization. This product wins by being constrained and clear.

## 5. Readability Over Density
Jira often fails because it looks like an administrative accident. Lattice must remain readable, visually calm, and easy to scan.

---

# Core User Experience Goals

The user should be able to do the following with minimal friction:

1. See what is not done, in progress, and completed
2. Create a work item with minimal steps (title-only creation in under 3 clicks)
3. Edit a work item without navigating through multiple obscure screens
4. Tag an item as blocked, delayed, or needs-review
5. Relate work items to one another
6. Group work beneath larger work items such as epics
7. Filter work meaningfully
8. Understand why something is blocked
9. Navigate from broad planning to specific implementation work
10. Use the system daily without feeling like they need training or a Jira priesthood

---

# Supported Platform

## Initial Platform
- Desktop-first responsive web application

## Secondary Support
- Tablet usable (layout must adapt at 768px breakpoint; touch targets at least 44px)
- Mobile view must be readable at 320px width without horizontal scrolling; tap targets must be at least 44px; full mobile optimization is not required in v1

---

# Technology Constraints

## Required Stack
- React
- TypeScript
- Tailwind CSS

## Recommended Tooling
- Vite
- React Router
- TanStack Query for server state
- Zustand or React Context for lightweight client state if needed
- React Hook Form or equivalent for forms
- Zod for validation
- A minimal headless component approach preferred over heavy UI frameworks

## Do Not Use
- Large enterprise UI frameworks that drag in unnecessary complexity
- Overengineered global state unless proven necessary
- Drag-and-drop libraries unless specifically required for board movement and implemented carefully

---

# Information Architecture

The frontend must provide a small, coherent navigation model.

## Primary Navigation
The application should have the following top-level areas:

1. Workspace Home
2. Work Board
3. Work List
4. Work Graph
5. Item Detail
6. Search / Filters
7. Settings

The user should not feel like these are totally separate products. They are different views over the same system.

---

# Core Domain Concepts the Frontend Must Represent

## Work Item
The primary object shown in the UI.

### Displayed Fields
- Title
- Description
- State
- Tags
- Type
- Parent
- Child items
- Relationships
- Created at
- Updated at

### Optional Later Fields
- Assignee
- Estimate
- Target date
- Linked branch or PR
- Activity history

## State
There are exactly three states (wire format values in parentheses):
- Not Done (`NotDone`)
- In Progress (`InProgress`)
- Completed (`Completed`)

The UI must never imply more lifecycle states than these three.

## Tags
Tags provide additional metadata and condition indicators.

Examples:
- blocked
- delayed
- needs-review
- high-priority
- backend
- frontend

Tags must remain visually and conceptually distinct from state.

## Type
Type is a lightweight label, not a workflow driver.

Examples:
- epic
- feature
- task
- bug
- spike

The UI must not imply type changes the lifecycle or rules of the item.

## Relationships
Relationships connect work items.

Initial relationship types:
- depends_on
- blocks
- relates_to
- duplicate_of

These relationships must be understandable in the UI without requiring the user to understand graph theory like a raccoon with a whiteboard.

## Parent / Child
Hierarchy must be supported, with one parent allowed per item.

Common usage:
- Epic → Feature
- Feature → Task
- Task → Subtask

The frontend must support hierarchy without requiring rigid type rules.

---

# Layout Requirements

## Global Application Shell
The shell should include:
- Left sidebar for major navigation
- Main content area
- Optional right-side contextual panel for filters, details, or relationship quick views
- Top bar with workspace name, search, and create button

## Visual Tone
- Clean
- Modern
- Sharp
- Slightly technical
- Not playful in a toy-like way
- No fake productivity “gamification”

## Design Characteristics
- Good whitespace
- Clear typography hierarchy
- Subtle borders and card structures
- Strong empty states
- Light visual chrome
- Data-dense where needed, but never claustrophobic

---

# Core Screens

## 1. Workspace Home

### Purpose
Provide a lightweight overview of work without becoming a metric cemetery.

### Must Include
- Summary counts by state
- Recently updated items
- Blocked items
- In-progress items
- My open items if assignees exist later
- Quick links to board, list, and graph views

### Must Not Become
- A giant dashboard with 47 charts no one trusts

### Acceptance Criteria
- Summary counts, blocked items, and in-progress items are visible without scrolling on a 1080p display
- Active work is reachable within 2 clicks from the home screen

---

## 2. Work Board View

### Purpose
Primary daily-use view for many users

### Model
Three columns only:
- Not Done
- In Progress
- Completed

### Card Contents
Each item card should show:
- Title
- Type
- Key tags
- Parent indicator if relevant
- Child indicator (icon only, no count — child counts are shown on detail view only in v1)
- Relationship warning icon if blocked or dependent
- Updated timestamp or relevant freshness hint

### Required Interactions
- Move item between columns
- Open item detail
- Quick edit tags
- Quick add item
- Filter board contents
- Group by parent optionally

### Optional Interactions
- Inline create child item
- Expand/collapse grouped hierarchies

### Board UX Rules
- Columns must be stable and easy to scan
- Card visuals must not overload the user
- Blocked items must be obvious without shouting
- Drag-and-drop must prevent backward state transitions for non-admin users (disable the drop target rather than allowing and snapping back)
- Admin users may drag backward; the UI sends the request with `override: true` and `X-Role: admin`

### Acceptance Criteria
- User can identify in-progress work immediately
- User can move an item from Not Done to In Progress with one direct interaction
- User can tell if an item is blocked without opening it

---

## 3. Work List View

### Purpose
Power-user and management-friendly view for scanning, filtering, sorting, and bulk understanding

### Required Columns
- Title
- State
- Type
- Tags
- Parent
- Relationship summary
- Updated at

### Optional Columns Later
- Assignee
- Created at
- Estimate
- Due date

### Required Features
- Search
- Sort
- Multi-filter
- Expand rows for more detail
- Bulk select
- Bulk tag
- Bulk state update if appropriate

### UX Rules
- Must remain readable
- No spreadsheet-from-hell experience
- Column density must be controlled

### Acceptance Criteria
- A filtered view of blocked bugs is reachable in 2 clicks (state + tag filter)
- User can filter by parent epic, type, state, and tags

---

## 4. Work Graph View

### Purpose
Show relationships and dependencies in a way that is useful, not decorative nonsense

### Primary Use Cases
- Understand blockers
- Understand dependency chains
- Understand how bugs relate to larger feature work
- Understand work clusters

### View Requirements
- Nodes represent work items
- Edges represent relationships
- Relationship type visually encoded
- Node state visually encoded
- Selected node reveals detail panel

### Graph Controls
- Zoom
- Pan
- Focus on selected node
- Show neighborhood only
- Filter by relationship type
- Toggle hierarchy vs dependency emphasis

### UX Rules
- Default graph should not dump the whole universe on screen
- Start from selected item or filtered subset
- Graph must be useful for answering actual work questions

### Acceptance Criteria
- User can see why an item is blocked
- User can see upstream and downstream dependencies
- User is not hit with an unreadable hairball by default

---

## 5. Item Detail View

### Purpose
Single source of truth for an individual work item

### Layout
Prefer a two-column detail layout or stacked sections with strong hierarchy.

### Required Sections
1. Header
   - Title
   - State
   - Type
   - Quick actions
2. Description
3. Tags
4. Parent / children
5. Relationships
6. Activity metadata
7. Related items preview

### Required Actions
- Edit title
- Edit description
- Change state
- Add/remove tags
- Change type
- Assign/change parent
- Add/remove relationships
- Create child item
- Navigate to related items

### UX Rules
- Editing must feel direct
- Relationship editing must not feel cryptic
- User must understand the difference between child items and related items

### Acceptance Criteria
- User can fully manage an item from this view
- User can inspect connected work without losing context

---

## 6. Search and Filtering Experience

### Purpose
Allow users to cut through work quickly

### Required Search
- Free text over title and description (client-side filtering within loaded page; backend text search parameter is a v2 feature)
- Tag search
- Type search
- State filter
- Parent filter
- Relationship-aware filters (`is_blocked`, `is_ready` query parameters are supported by the backend)

### Required Filter Controls
- State
- Type
- Tags
- Parent
- Has dependencies
- Is blocked
- Has children

### UX Rules
- Filtering must be consistent across board, list, and graph views
- Saved filter presets are desirable later but not required in v1
- Search results must update within 200ms of keystroke (matches P95 target)

### Acceptance Criteria
- User can quickly isolate blocked in-progress bugs under a given epic
- Filters can be reset easily
- Users understand current active filters

---

# Interaction Requirements

## Creating a Work Item

### Entry Points
- Global create button
- Board inline create
- Child create from item detail
- Quick create from home

### Form Fields
Required:
- Title

Optional:
- Description
- Type
- Tags
- Parent
- Relationships

### UX Requirements
- Quick create path must be minimal
- Advanced fields (Description, Type, Tags, Parent, Relationships) are progressively disclosed behind an "Add details" toggle
- User should not be forced to think about every field up front

### Acceptance Criteria
- User can create a basic item with only title in seconds
- User can create a child item from a parent context without re-entering context manually

---

## Editing a Work Item

### Required Behaviors
- Inline edit for title, description, tags, state, and type fields
- Autosave all field changes with a 500ms debounce and show a "Saved" indicator in the item header. On failure, show a "Failed to save" indicator with a manual retry button; keep the user's draft value in the field (do not revert) and highlight the field as unsaved
- Field validation
- Optimistic UI where safe

### Acceptance Criteria
- Edits feel immediate
- Errors are clear and recoverable

---

## State Changes

### State Transition Matrix

| From | To | User | Admin |
|------|----|------|-------|
| NotDone | InProgress | Allowed | Allowed |
| InProgress | Completed | Allowed | Allowed |
| NotDone | Completed | **Rejected** (must go through InProgress) | **Rejected** |
| Completed | InProgress | Rejected | Allowed (with override) |
| InProgress | NotDone | Rejected | Allowed (with override) |
| Completed | NotDone | Rejected | Rejected (must step back one at a time) |

### Backward Transitions
- Backward state transitions (e.g., Completed → In Progress, In Progress → Not Done) are supported only via admin override
- The backend rejects backward transitions unless the request includes `override: true` and the `X-Role: admin` header
- The UI must disable backward transition controls for non-admin users (disabled buttons, non-droppable board columns)
- Admin users see a clearly labeled "Override" action that is visually distinct from normal state changes
- The frontend reads the user's role from a `window.__LATTICE_CONFIG__.role` global set by the hosting page or API gateway reverse proxy. The value is either `"admin"` or `"user"` (default). The frontend sends `X-Role: admin` on requests only when this value is `"admin"`

### UX Requirements
- State change must be simple
- The UI must not present fake workflow options

---

## Tag Management

### Required Features
- Add tag
- Remove tag
- Typeahead for existing tags
- Create new tags freely

### UX Rules
- Tags should be easy to scan
- Condition tags such as blocked should have stronger visual treatment than generic topical tags

### Acceptance Criteria
- User can tag an item blocked in one quick action
- Existing tags are discoverable

---

## Relationship Management

### Required Features
- Add relationship from item detail
- Remove relationship
- View reverse relationships
- Relationship labels shown in human-readable language

### Human-Readable Examples
Instead of only:
- depends_on

Prefer:
- Depends on: API authentication task

Instead of only:
- blocks

Prefer:
- Blocks: OAuth setup

### UX Rules
- Relationship creation must feel clear
- The user must know which direction the relationship goes
- The UI must reduce mistakes in directional linking

### Acceptance Criteria
- User can relate a bug to a feature
- User can define dependency chains
- User can understand why an item is considered blocked

---

## Parent/Child Management

### Required Features
- Set parent
- Remove parent
- View children
- Create child directly from parent

### UX Rules
- Parent-child must feel like grouping and decomposition, not like bureaucracy
- Child items should be visible from the parent detail view and optionally in list/board grouping

### Delete Behavior
- Deleting an item that has children must show a confirmation dialog: "This item has N children. Deleting it will orphan them (remove their parent). Continue?"
- The backend handles cascade: relationships and tags are deleted automatically, children's parent_id is set to null

### Acceptance Criteria
- User can create an epic with child tasks
- User can re-parent an item if needed

---

# Frontend State and Data Flow

## Server State
Use TanStack Query or equivalent for:
- Fetching work items
- Fetching single item details
- Mutations for create/update/delete
- Relationship operations

## Local UI State
Local state may be used for:
- Modal open/close
- Draft forms
- Local filters
- Graph selection
- Column UI settings

## Caching Rules
- List and detail views should stay synchronized after mutations
- Query invalidation must be deliberate, not brute-force chaos
- Optimistic updates should be used where latency would hurt UX

---

# Routing

## Required Routes
- `/`
- `/board`
- `/list`
- `/graph`
- `/items/:id`
- `/settings`

### Optional Routes Later
- `/epics/:id`
- `/search`
- `/views/:savedViewId`

### Routing Requirements
- Deep links must work for item detail pages
- Browser navigation must behave normally
- Selected filters in list/board should ideally be URL-representable when practical

---

# Components

## Core Shared Components
The frontend should include reusable components for:

- AppShell
- SidebarNav
- Topbar
- PageHeader
- WorkItemCard
- WorkItemRow
- WorkItemStateBadge
- TagBadge
- TypeBadge
- RelationshipList
- ParentChildPanel
- SearchInput
- FilterPanel
- EmptyState
- ErrorState
- LoadingState
- Modal
- Drawer
- ConfirmDialog
- InlineEditableText
- GraphCanvas
- BoardColumn

---

# Form Requirements

## General Rules
- Forms must validate cleanly
- Errors must be visible near fields
- Save state must be visible
- Cancel must not be destructive

## Validation Examples
- Title required
- Title must be between 1 and 500 characters (matches backend constraint)
- Parent cannot be self
- Parent hierarchy must be acyclic (the backend rejects circular parent chains; the frontend should show the backend's validation error clearly)
- Relationship target cannot be self
- Relationship duplicates prevented where possible

---

# Accessibility Requirements

The frontend must be accessible enough to be used professionally and responsibly.

## Required
- Keyboard navigation
- Proper focus handling
- Semantic HTML
- Labels for controls
- Color contrast sufficient for professional usage
- Visible focus states
- Non-color-only state indicators

## Acceptance Criteria
- User can navigate key workflows without a mouse
- State, tags, and warnings are understandable without relying only on color

---

# Performance Requirements

## Perceived Performance Goals (P95 targets, desktop with broadband ≥10 Mbps)
- Board, list, and home views must render within 300ms of navigation
- Search/filter updates must reflect within 200ms of user input
- Detail view transitions must complete within 250ms
- Item creation (title only) must complete round-trip within 500ms

## Technical Goals
- Route-level code splitting (each page loaded on demand)
- Avoid unnecessarily rerendering large lists
- Virtualization may be needed for large list views
- Graph rendering must be scoped for performance

## Non-Goals for v1
- Infinite scale to giant enterprise datasets in browser
- Perfect graph rendering for thousands of nodes at once

---

# Visual Semantics

## State Representation
State must be visually distinct and consistent.

Suggested semantics:
- Not Done: neutral
- In Progress: emphasized
- Completed: subdued but clearly done

Do not assign state semantics that imply more workflow steps than exist.

## Tag Representation
Tags should be pill-like and scan-friendly.

The following condition tags require stronger visual emphasis (bold pill, warning color):
- `blocked`
- `delayed`
- `needs-review`

All other tags use the default neutral pill style.

## Relationship Representation
Relationships should be human-readable and visually secondary until relevant.

## Hierarchy Representation
Parent-child grouping should be clear through indentation, chips, panels, or grouped sections.

---

# Empty States

Every major screen must have strong empty states.

## Examples
- No work items yet
- No results for current filters
- No relationships yet
- No child items yet

Each empty state should:
- Explain what is missing
- Suggest the next useful action
- Never feel dead or broken

---

# Error Handling

## Required
- API errors displayed clearly
- Inline mutation failures surfaced to the user
- Retry options where sensible
- Global catastrophic failure state if app cannot initialize

## UX Rules
- Errors must be understandable
- Do not dump raw backend garbage into the UI unless necessary for debugging mode
- Users must know whether an action succeeded or failed

---

# Notifications and Feedback

## Required
- Success feedback for create/update/delete actions
- Clear save confirmation
- Non-obnoxious toast or inline confirmation patterns

## UX Rules
- Keep notifications restrained
- Avoid constant pop-up spam
- Feedback should reassure, not nag

---

# API Contract

The backend API is fully implemented. The frontend must use these exact endpoints and wire formats. All request/response bodies are `application/json`.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST /workitems` | Create work item (state always starts as `NotDone`) |
| `GET /workitems/{id}` | Get work item with tags and relationships |
| `PATCH /workitems/{id}` | Partial update (only provided fields change) |
| `DELETE /workitems/{id}` | Delete work item (cascades relationships and tags) |
| `GET /workitems` | List with filtering and pagination |
| `POST /workitems/{id}/relationships` | Add relationship |
| `DELETE /workitems/{id}/relationships/{rel_id}` | Remove relationship |
| `GET /workitems/{id}/cycles` | Detect dependency cycles (called on-demand from Graph view and after adding a `depends_on`/`blocks` relationship) |

## List Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `state` | string | `NotDone`, `InProgress`, or `Completed` |
| `tags` | string | Comma-separated; AND logic |
| `type` | string | Filter by type |
| `parent_id` | UUID | Filter by parent |
| `is_blocked` | bool | Has unresolved `depends_on` |
| `is_ready` | bool | All `depends_on` targets are `Completed` |
| `page` | int | Default: 1 |
| `page_size` | int | Default: 50, max: 200 |

Note: `is_blocked` and `is_ready` cannot be used together. The UI must disable the `is_ready` filter when `is_blocked` is active, and vice versa.

## Wire Format — WorkItem

```json
{
  "id": "UUID",
  "title": "string (required, max 500 chars)",
  "description": "string (max 10000 chars)",
  "state": "NotDone | InProgress | Completed",
  "tags": ["string"],
  "type": "string | omitted",
  "parent_id": "UUID | null (GET/POST) or empty string to unset (PATCH)",
  "relationships": [
    { "id": "UUID", "type": "blocks | depends_on | relates_to | duplicate_of", "target_id": "UUID" }
  ],
  "created_at": "RFC3339",
  "updated_at": "RFC3339",
  "is_blocked": "boolean (derived by backend: true if any depends_on target is not Completed)"
}
```

## Wire Format — List Response

```json
{
  "items": [WorkItem, ...],
  "total": 42,
  "page": 1,
  "page_size": 50
}
```

## Wire Format — Error Response

```json
{
  "error": { "code": "NOT_FOUND | INVALID_INPUT | FORBIDDEN | INVALID_TRANSITION | VALIDATION_ERROR", "message": "string" }
}
```

## Error Code → UI Behavior Mapping

| HTTP Status | Code | UI Behavior |
|-------------|------|-------------|
| 400 | `INVALID_INPUT` | Show inline validation error near the relevant field |
| 403 | `FORBIDDEN` | Show toast: "Admin permission required" |
| 404 | `NOT_FOUND` | Navigate to a "not found" empty state |
| 409 | `INVALID_TRANSITION` | Show toast explaining the transition is not allowed |
| 422 | `VALIDATION_ERROR` | Show inline error (e.g., cycle detected, depth exceeded) |
| 5xx | — | Show global error banner with retry option |

## PATCH Semantics

- Only included fields are modified; omitted fields are unchanged
- To unset `parent_id`, send `"parent_id": ""`
- Tags are replaced entirely if present in the request; omit `tags` to leave unchanged

## Derived Blocked Status

The `is_blocked` status is computed by the backend. The frontend queries it via the `is_blocked=true` list filter parameter. The frontend does not need to compute blocked status from relationships.

The frontend should abstract API access behind a client layer so backend changes do not infect UI code everywhere.

---

# Suggested Frontend Directory Structure

```text
src/
  app/
    router/
    providers/
    layouts/
  components/
    common/
    workitems/
    board/
    list/
    graph/
    forms/
  features/
    workitems/
      api/
      hooks/
      components/
      types/
      utils/
    filters/
    relationships/
    hierarchy/
  pages/
    HomePage.tsx
    BoardPage.tsx
    ListPage.tsx
    GraphPage.tsx
    ItemDetailPage.tsx
    SettingsPage.tsx
  lib/
    api/
    utils/
    constants/
  styles/
  main.tsx


⸻

Required Derived UI Behaviors

Blocked Indicator

An item appears blocked when the backend `is_blocked` list filter returns it (meaning it has unresolved `depends_on` dependencies). The `blocked` tag is a separate user-applied label for manual flagging and does not affect the `is_blocked` computed status.

The "Is blocked" filter in the UI maps to the `is_blocked=true` query parameter. To find items tagged `blocked`, use the tags filter. The frontend does not compute blocked status from relationships.

Child Summary

An item with children should show:
- Child count on the detail view only (queried via `GET /workitems?parent_id={id}&page_size=1` to get `total`)
- Board and list views do not display child counts in v1 to avoid N+1 queries

Relationship Summary

List and board views show a compact relationship summary derived from the `relationships` array:

| Relationship type | Summary label |
|-------------------|---------------|
| `depends_on` | "N dependencies" |
| `blocks` | "N blockers" |
| `relates_to` | "N related" |
| `duplicate_of` | "N duplicates" |

⸻

Feature Scope for v1

Must Have
	•	App shell
	•	Board view
	•	List view
	•	Graph view
	•	Item detail view
	•	Create/edit work item
	•	Tag editing
	•	Relationship editing
	•	Parent-child handling
	•	Filtering/search
	•	Error/loading states
	•	Responsive desktop web UX

Nice to Have
	•	Inline child creation
	•	URL-persisted filters
	•	Keyboard shortcuts
	•	Bulk actions
	•	Saved filter presets

Explicitly Out of Scope for v1
	•	Custom workflows
	•	Highly configurable dashboards
	•	Sprint management
	•	Story points obsession
	•	Plugin ecosystem
	•	Time tracking bureaucracy
	•	Complex permissions matrix
	•	Chat/social features
	•	Native mobile apps
	•	Realtime collaborative editing
	•	AI-generated work suggestions

⸻

UX Anti-Requirements

The frontend must not become the following:
	1.	A clone of Jira with cleaner colors
	2.	A project management toy for people who never shipped software
	3.	A dashboard circus
	4.	A kanban-only system that fails the moment dependencies matter
	5.	A graph demo that looks smart and helps nobody
	6.	A form-heavy enterprise punishment engine

⸻

Acceptance Criteria

Global
	1.	The application builds and runs cleanly
	2.	Navigation is clear and consistent
	3.	The main workflows can be used without training documentation

Board
	1.	User can view items by the three states
	2.	User can move an item between states
	3.	User can identify blocked items quickly

List
	1.	User can search and filter effectively
	2.	User can inspect hierarchy and relationships in compact form

Graph
	1.	User can inspect dependencies between related work items
	2.	Graph starts from focused context instead of dumping all data blindly

Detail
	1.	User can fully manage a work item from the detail screen
	2.	User can add and remove relationships
	3.	User can manage parent-child links

Create/Edit
	1.	User can create an item with only a title
	2.	User can progressively add metadata
	3.	Validation errors are clear

Quality
	1.	The UI is visually calm and professional
	2.	State, tags, and relationships are clearly differentiated
	3.	Accessibility basics are handled
	4.	Performance meets P95 latency targets defined in the Performance Requirements section

⸻

Build Guidance for the Agent

Important Implementation Guidance
	1.	Start with the core shell and routing
	2.	Implement shared domain types and API client early
	3.	Build item detail and create/edit flows before polishing graph complexity
	4.	Keep board/list/detail in sync through shared hooks and query keys
	5.	Avoid overbuilding settings or customization
	6.	Build graph view as focused and filtered, not universal and chaotic
	7.	Prefer composable, testable components
	8.	Keep visual language consistent across all views

Suggested Build Order
	1.	App shell and routing
	2.	API client and domain types
	3.	Board view
	4.	Item detail page
	5.	Create/edit flows
	6.	List view and filtering
	7.	Relationship management UI
	8.	Graph view
	9.	Accessibility pass
	10.	Performance pass
	11.	Polish empty/error states

⸻

Future Extensions

The frontend should be designed so the following can be added later without structural rewrite:
	•	Assignees
	•	Dates and scheduling
	•	Git integration
	•	AI-assisted suggestions
	•	Saved views
	•	Notifications center
	•	Activity history
	•	Team-specific workspaces
	•	Permissions model
	•	Rich reporting

The architecture should remain clean enough that future expansion does not require ripping out half the app like an angry plumber.

⸻

Final Guiding Constraint

The frontend must make complex development work feel understandable without making the user manage system complexity directly.

If a design decision introduces more visible complexity to solve an internal modeling problem, that design decision is wrong unless there is overwhelming evidence otherwise.
