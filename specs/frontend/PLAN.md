# PLAN.md

# Lattice Frontend — Phased Implementation Plan

## Technology Stack

- **Framework:** React 19 + TypeScript 5.x
- **Build Tool:** Vite
- **Styling:** Tailwind CSS 4
- **Routing:** React Router 7
- **Server State:** TanStack Query v5
- **Forms:** React Hook Form + Zod
- **Graph Rendering:** @xyflow/react (React Flow)
- **Drag-and-Drop:** @dnd-kit (board column movement)
- **Testing:** Vitest + React Testing Library
- **Linting:** ESLint + Prettier

## Project Layout

```
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.ts
├── src/
│   ├── main.tsx                      # React root + providers
│   ├── app/
│   │   ├── App.tsx                   # Root component
│   │   ├── router.tsx                # Route definitions
│   │   ├── providers.tsx             # QueryClient, config context
│   │   └── layouts/
│   │       └── AppShell.tsx          # Sidebar + topbar + main content
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts            # Base fetch wrapper, error handling
│   │   │   ├── workitems.ts         # WorkItem CRUD functions
│   │   │   └── relationships.ts     # Relationship add/remove
│   │   ├── types.ts                 # WorkItem, Relationship, State, etc.
│   │   ├── config.ts                # Read window.__LATTICE_CONFIG__
│   │   ├── constants.ts             # Condition tags, state colors, etc.
│   │   └── validation.ts            # Zod schemas for forms
│   ├── hooks/
│   │   ├── useWorkItems.ts          # TanStack Query hooks for list/get
│   │   ├── useWorkItemMutations.ts  # Create/update/delete mutations
│   │   ├── useRelationships.ts      # Add/remove relationship mutations
│   │   ├── useCycles.ts             # Cycle detection query hook
│   │   ├── useFilters.ts            # Filter state + URL sync
│   │   └── useConfig.ts             # Read role, admin status
│   ├── components/
│   │   ├── common/
│   │   │   ├── EmptyState.tsx
│   │   │   ├── ErrorState.tsx
│   │   │   ├── LoadingState.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── ConfirmDialog.tsx
│   │   │   ├── Toast.tsx
│   │   │   └── InlineEditableText.tsx
│   │   ├── workitems/
│   │   │   ├── WorkItemCard.tsx      # Board card
│   │   │   ├── WorkItemRow.tsx       # List row
│   │   │   ├── StateBadge.tsx
│   │   │   ├── TagBadge.tsx
│   │   │   ├── TypeBadge.tsx
│   │   │   └── RelationshipSummary.tsx
│   │   ├── filters/
│   │   │   ├── FilterPanel.tsx
│   │   │   ├── SearchInput.tsx
│   │   │   └── FilterChips.tsx
│   │   └── forms/
│   │       ├── CreateWorkItemForm.tsx
│   │       ├── EditWorkItemFields.tsx
│   │       ├── TagEditor.tsx
│   │       ├── RelationshipEditor.tsx
│   │       └── ParentPicker.tsx
│   ├── pages/
│   │   ├── HomePage.tsx
│   │   ├── BoardPage.tsx
│   │   ├── ListPage.tsx
│   │   ├── GraphPage.tsx
│   │   ├── ItemDetailPage.tsx
│   │   ├── SettingsPage.tsx
│   │   └── NotFoundPage.tsx
│   └── styles/
│       └── index.css                 # Tailwind directives + custom tokens
```

---

## Phase 1 — Project Scaffolding, Types, and API Client

### Goal
Establish the project skeleton, domain types, API client layer, and app shell. Everything after this phase builds on working routing and data fetching.

### Tasks

1. **Initialize project**
   - `npm create vite@latest frontend -- --template react-ts`
   - Install dependencies: `react-router`, `@tanstack/react-query`, `tailwindcss`, `zod`, `react-hook-form`
   - Configure Vite proxy to forward `/workitems` to backend (`http://localhost:8080`)

2. **Define domain types** (`src/lib/types.ts`)
   - `WorkItemState = "NotDone" | "InProgress" | "Completed"`
   - `RelationshipType = "depends_on" | "blocks" | "relates_to" | "duplicate_of"`
   - `Relationship { id, type, target_id }`
   - `WorkItem { id, title, description, state, tags, type, parent_id, relationships, created_at, updated_at, is_blocked }`
   - `ListResponse { items, total, page, page_size }`
   - `ApiError { error: { code, message } }`
   - `ListFilter` type matching all query parameters

3. **Build API client** (`src/lib/api/client.ts`)
   - Base `apiFetch<T>(path, options)` — sets Content-Type, reads `window.__LATTICE_CONFIG__` for role, attaches `X-Role` header when admin
   - Parses error responses into typed `ApiError`, throws on non-2xx
   - `src/lib/api/workitems.ts`: `createWorkItem`, `getWorkItem`, `updateWorkItem`, `deleteWorkItem`, `listWorkItems`
   - `src/lib/api/relationships.ts`: `addRelationship`, `removeRelationship`
   - `src/lib/api/cycles.ts`: `detectCycles`

4. **Config reader** (`src/lib/config.ts`)
   - Read `window.__LATTICE_CONFIG__` with type assertion
   - Export `getRole(): "admin" | "user"` and `isAdmin(): boolean`

5. **Zod validation schemas** (`src/lib/validation.ts`)
   - `createWorkItemSchema`: title required, 1-500 chars; description optional, max 10000 chars
   - `updateWorkItemSchema`: all optional, title 1-500 if present, description max 10000 if present
   - `addRelationshipSchema`: type must be valid, target_id required

6. **App shell and routing**
   - `src/app/providers.tsx`: wrap `QueryClientProvider` + `BrowserRouter`
   - `src/app/router.tsx`: routes for `/`, `/board`, `/list`, `/graph`, `/items/:id`, `/settings`, `*` → NotFoundPage
   - `src/app/layouts/AppShell.tsx`: left sidebar nav (Home, Board, List, Graph, Settings) + top bar (workspace name, global create button, search input placeholder)
   - Stub pages: each page renders its name + the shell

7. **Tailwind setup** (`src/styles/index.css`)
   - Define color tokens for states: neutral (NotDone), blue/emphasized (InProgress), green-subdued (Completed)
   - Define condition tag colors (blocked=red, delayed=amber, needs-review=purple)
   - Base typography scale

### Deliverables
- `npm run dev` renders the app shell with working nav between all routes
- API client functions are importable and typed
- No runtime errors, TypeScript strict mode passes

### Tests
- Unit test for `apiFetch` error parsing (mock fetch)
- Unit test for Zod schemas (valid/invalid inputs)
- Unit test for config reader (with/without global)

---

## Phase 2 — Board View

### Goal
Deliver the primary daily-use view: a three-column kanban board with drag-and-drop state transitions, work item cards, and basic filtering.

### Dependencies
- Phase 1 (types, API client, shell, routing)

### Tasks

1. **TanStack Query hooks** (`src/hooks/`)
   - `useWorkItems(filter)` — calls `listWorkItems`, returns `{ data, isLoading, error }`
   - `useWorkItemMutations()` — `createMutation`, `updateMutation`, `deleteMutation` with query invalidation on success
   - Invalidation strategy: on mutate success, invalidate `["workitems"]` list queries and `["workitem", id]` detail queries

2. **WorkItemCard component** (`src/components/workitems/WorkItemCard.tsx`)
   - Displays: title, type badge, key tags (max 3 + overflow count), parent indicator chip, child icon (if has children via parent_id filter), blocked/dependency warning icon, relative updated time
   - Click opens `/items/:id`
   - Condition tags (`blocked`, `delayed`, `needs-review`) use warning pill style

3. **StateBadge, TagBadge, TypeBadge** — small shared components
   - StateBadge: colored dot + label, 3 states only
   - TagBadge: pill with conditional emphasis for condition tags
   - TypeBadge: neutral pill (epic, feature, task, bug, spike)

4. **BoardColumn component**
   - Renders column header with state name + count
   - Scrollable card list
   - Drop zone for drag-and-drop
   - "Quick add" input at bottom of NotDone column

5. **BoardPage** (`src/pages/BoardPage.tsx`)
   - Three columns: NotDone, InProgress, Completed
   - Fetches all items (paginated, initial page_size=200 to load board fully)
   - Groups items by state into columns

6. **Drag-and-drop** using `@dnd-kit`
   - Dragging a card to a different column triggers `updateMutation` with new state
   - Forward one-step transitions (NotDone→InProgress, InProgress→Completed) allowed for all users
   - NotDone→Completed is rejected for all users (drop target disabled on Completed column when dragging from NotDone)
   - Backward transitions: drop target disabled for non-admin users (column shows visual "not allowed" indicator on hover)
   - Admin users: backward drop sends `override: true` + `X-Role: admin`
   - Optimistic update: card moves immediately, rolls back on API error with toast

7. **Quick add** — input at bottom of NotDone column
   - Title-only creation, Enter to submit
   - Calls `createMutation`, card appears optimistically

8. **Empty states** — each column shows empty state when no items

### Deliverables
- Board renders three columns with real data from backend
- Drag-and-drop moves items between states
- Quick add creates items
- Admin/non-admin distinction enforced on backward transitions

### Tests
- Unit: WorkItemCard renders title, tags, badges
- Unit: BoardColumn renders items, empty state
- Integration: drag from NotDone to InProgress calls updateMutation with correct state

---

## Phase 3 — Item Detail Page + Create/Edit

### Goal
Full item management from a single page: view, inline edit all fields, autosave, manage tags, view parent/children, view relationships.

### Dependencies
- Phase 2 (hooks, card components, mutations)

### Tasks

1. **useWorkItem(id) hook** — fetches single item via `getWorkItem`, returns `{ data, isLoading, error }`

2. **ItemDetailPage** (`src/pages/ItemDetailPage.tsx`)
   - Two-column layout: main content (left), metadata sidebar (right)
   - Left: title (inline editable), description (inline editable textarea), relationships list, related items preview
   - Right: state selector, type selector, tags editor, parent/children panel, timestamps

3. **InlineEditableText component** (`src/components/common/InlineEditableText.tsx`)
   - Click to edit, blur or Enter to save
   - 500ms debounced autosave via `updateMutation`
   - Shows "Saved" indicator on success, "Failed to save" + retry on error
   - Keeps draft value on failure, highlights field as unsaved

4. **State selector**
   - Dropdown or button group showing allowed transitions from current state
   - Non-admin: only forward transitions shown as enabled
   - Admin: backward transitions shown with "Override" label
   - Calls `updateMutation` on selection

5. **TagEditor component** (`src/components/forms/TagEditor.tsx`)
   - Displays current tags as removable pills
   - Input with typeahead (sourced from tags seen in recent list queries)
   - Enter to add new tag
   - Autosaves: sends full tag array via PATCH on each add/remove

6. **ParentChildPanel** (`src/components/forms/ParentPicker.tsx` + panel)
   - Shows current parent as clickable link (or "None")
   - "Change parent" opens a searchable item picker modal
   - "Remove parent" sends PATCH with `parent_id: ""`
   - Children section: queries `GET /workitems?parent_id={id}&page_size=10`, shows list with links
   - "Create child" button opens quick create with parent pre-filled

7. **CreateWorkItemForm** (`src/components/forms/CreateWorkItemForm.tsx`)
   - Modal triggered by global "Create" button or board quick-add expansion
   - Title required, "Add details" toggle reveals: description, type, tags, parent
   - Zod validation, error messages inline
   - On submit: `createMutation`, close modal, navigate to new item or stay

8. **Delete with confirmation**
   - Delete button on detail page
   - If item has children (checked via `GET /workitems?parent_id={id}&page_size=1`), show ConfirmDialog: "This item has N children. Deleting it will orphan them."
   - On confirm: `deleteMutation`, navigate to board

### Deliverables
- Full item CRUD from detail page
- Inline editing with autosave
- Tag management with typeahead
- Parent/child management
- Create modal with progressive disclosure

### Tests
- Unit: InlineEditableText debounce + save indicator
- Unit: TagEditor add/remove
- Unit: CreateWorkItemForm validation
- Unit: State selector shows correct transitions for user vs admin

---

## Phase 4 — List View + Filtering

### Goal
Power-user list view with sortable columns, multi-filter panel, client-side text search, and consistent filter controls shared with board.

### Dependencies
- Phase 3 (hooks, mutation patterns, shared components)

### Tasks

1. **useFilters hook** (`src/hooks/useFilters.ts`)
   - Manages filter state: `{ state, type, tags, parent_id, is_blocked, is_ready }`
   - Syncs filter state to URL search params (e.g., `?state=InProgress&tags=urgent`)
   - `is_blocked` and `is_ready` are mutually exclusive: setting one clears the other
   - Returns `{ filters, setFilter, clearFilters, activeFilterCount }`

2. **FilterPanel component** (`src/components/filters/FilterPanel.tsx`)
   - State filter: three toggle buttons (NotDone, InProgress, Completed)
   - Type filter: dropdown populated from observed types
   - Tags filter: multi-select pills with typeahead
   - Parent filter: searchable item picker
   - Is blocked / Is ready: toggle (mutually exclusive)
   - Active filter chips shown above results with "Clear all" button
   - Shared between Board and List pages

3. **SearchInput component** (`src/components/filters/SearchInput.tsx`)
   - Debounced text input (200ms)
   - Client-side filter: filters `items` array by title/description substring match
   - Shown in top bar and in list/board views

4. **WorkItemRow component** (`src/components/workitems/WorkItemRow.tsx`)
   - Table row with columns: Title (link), State (badge), Type (badge), Tags (pills, max 3), Parent (link), Relationship summary (e.g., "2 deps, 1 blocker"), Updated at (relative)
   - Expandable: click row to show description + full relationship list inline

5. **ListPage** (`src/pages/ListPage.tsx`)
   - Table layout with sortable column headers (sort by title, state, type, updated_at)
   - Client-side sorting on the loaded page
   - Pagination controls: page size selector (20/50/100/200), prev/next buttons
   - FilterPanel sidebar or top section
   - Empty state when no results match

6. **Integrate FilterPanel into BoardPage**
   - Board uses the same `useFilters` hook
   - Filter panel toggles open from a "Filter" button
   - Board columns show filtered items

7. **RelationshipSummary component** (`src/components/workitems/RelationshipSummary.tsx`)
   - Compact inline display: "2 dependencies · 1 blocker"
   - Counts per relationship type from the `relationships` array
   - Used in WorkItemRow and WorkItemCard

### Deliverables
- List view with all columns, sorting, pagination
- Filter panel shared between board and list
- URL-persisted filters
- Client-side text search

### Tests
- Unit: useFilters sets/clears filters, URL sync
- Unit: FilterPanel renders controls, mutual exclusion of is_blocked/is_ready
- Unit: WorkItemRow renders all columns
- Unit: SearchInput debounce behavior

---

## Phase 5 — Relationship Management UI

### Goal
Users can add, remove, and understand relationships from the item detail page. Human-readable labels, directional clarity, and reverse relationship visibility.

### Dependencies
- Phase 3 (detail page, mutation hooks)

### Tasks

1. **useRelationships hook** (`src/hooks/useRelationships.ts`)
   - `addRelationshipMutation(sourceId, { type, target_id })` — calls `addRelationship`, invalidates source item query
   - `removeRelationshipMutation(sourceId, relId)` — calls `removeRelationship`, invalidates source item query
   - After adding `depends_on` or `blocks`, calls `detectCycles` and shows warning toast if cycles found

2. **RelationshipList component** (in detail page)
   - Groups relationships by type
   - Each relationship row: human-readable label ("Depends on: {target title}"), link to target, remove button
   - Human-readable mapping: `depends_on` → "Depends on", `blocks` → "Blocks", `relates_to` → "Related to", `duplicate_of` → "Duplicate of"

3. **RelationshipEditor component** (`src/components/forms/RelationshipEditor.tsx`)
   - "Add relationship" button opens inline form
   - Type selector dropdown (4 types)
   - Target item picker: searchable dropdown querying `listWorkItems`
   - Direction guidance: shows sentence preview ("This item depends on: {selected target}")
   - Prevents selecting self as target
   - On submit: `addRelationshipMutation`

4. **Reverse relationships** (read-only section on detail page)
   - Query: for each item shown, check incoming relationships by searching `listWorkItems` with relationship filters
   - Display section: "Blocked by" / "Depended on by" / "Related from" / "Duplicate from"
   - This is informational only — removal is done from the source item

5. **Cycle detection integration**
   - After adding a `depends_on` or `blocks` relationship, fetch `GET /workitems/{id}/cycles`
   - If cycles returned: show warning toast "Dependency cycle detected: A → B → A" with link to graph view
   - Non-blocking — cycles are allowed per spec, just surfaced

### Deliverables
- Add/remove relationships from detail page
- Human-readable relationship labels with directional clarity
- Reverse relationship visibility
- Cycle detection warning after dependency additions

### Tests
- Unit: RelationshipList renders grouped relationships with labels
- Unit: RelationshipEditor prevents self-link, validates type
- Unit: Cycle warning toast shown when cycles detected
- Integration: add relationship → query invalidation → list refreshes

---

## Phase 6 — Graph View

### Goal
Interactive dependency graph starting from a focused item or filtered subset. Nodes represent work items, edges represent relationships, with state and type visually encoded.

### Dependencies
- Phase 5 (relationship data, cycle detection)

### Tasks

1. **Install @xyflow/react** (React Flow)
   - Lightweight, handles zoom, pan, node selection, edge rendering

2. **GraphPage** (`src/pages/GraphPage.tsx`)
   - Entry: URL param `?focus={id}` or starts from a filtered set
   - If `focus` param: load the focused item + its relationships, then load each related item (1 hop by default)
   - If no focus: show a picker ("Select an item to explore its dependencies")
   - Controls sidebar: relationship type filter toggles, depth slider (1-3 hops), "Show hierarchy" toggle

3. **Graph node component**
   - Custom React Flow node
   - Shows: title (truncated), state badge (color-coded), type badge, blocked indicator
   - Click selects node → detail panel opens on right
   - Selected node highlighted

4. **Graph edge component**
   - Edge label: relationship type
   - Style by type: `depends_on`/`blocks` = solid directional arrow, `relates_to` = dashed, `duplicate_of` = dotted
   - Color by target state: completed edges subdued, blocked edges emphasized

5. **Graph layout**
   - Use dagre layout algorithm via React Flow layout helpers (lighter weight, standard React Flow integration)
   - Top-to-bottom for dependency chains, left-to-right for hierarchy
   - Auto-fit on initial load

6. **Detail panel on node select**
   - Right sidebar shows: title, state, type, tags, description snippet, relationship list
   - "Open full detail" link to `/items/:id`
   - "Focus here" button re-centers graph on selected node

7. **Cycle visualization**
   - On graph load, call `detectCycles` for the focused item
   - If cycles exist, highlight cycle edges in red with animated dash pattern
   - Show banner: "Dependency cycle detected" with cycle path

### Deliverables
- Interactive graph starting from focused item
- Zoom, pan, node selection
- Relationship type filtering
- Cycle visualization
- Detail panel on select

### Tests
- Unit: Graph node renders title, state badge, blocked indicator
- Unit: Graph edge styles match relationship type
- Integration: focus param loads correct neighborhood

---

## Phase 7 — Home Page + Workspace Overview

### Goal
Lightweight home screen showing work status at a glance, quick navigation to active work, and entry points for creating items.

### Dependencies
- Phase 4 (list hooks, filter hooks)

### Tasks

1. **HomePage** (`src/pages/HomePage.tsx`)
   - Summary cards row: count of NotDone, InProgress, Completed items (3 separate `listWorkItems` calls with `page_size=1` to get `total`)
   - "Blocked Items" section: `listWorkItems({ is_blocked: true, page_size: 5 })`, show as compact card list
   - "In Progress" section: `listWorkItems({ state: "InProgress", page_size: 5 })`, compact card list
   - "Recently Updated" section: `listWorkItems({ page_size: 5 })` (default sort is by updated_at desc), compact card list
   - Quick navigation links: "View Board", "View List", "View Graph"
   - "Create Work Item" prominent button

2. **Compact card list component** — reusable mini card for home sections
   - Title, state badge, type badge, blocked indicator
   - Click navigates to `/items/:id`
   - "View all →" link at bottom navigates to list with appropriate filter

3. **Layout: all sections visible without scrolling on 1080p** (per spec acceptance criteria)

### Deliverables
- Home page with state summary, blocked items, in-progress items, recent items
- Quick navigation to board/list/graph
- Create button

### Tests
- Unit: HomePage renders summary counts
- Unit: Blocked items section queries with correct filter
- Unit: Compact card renders title + badges

---

## Phase 8 — Settings, Empty/Error States, and Toast System

### Goal
Complete the shell: settings page, polished empty states for all views, global error handling, and toast notification system.

### Dependencies
- All previous phases

### Tasks

1. **SettingsPage** (`src/pages/SettingsPage.tsx`)
   - Display current role (from config)
   - Display backend URL
   - Minimal — no complex configuration per spec philosophy

2. **Toast system** (`src/components/common/Toast.tsx`)
   - Global toast container rendered in AppShell
   - Types: success (green), error (red), warning (amber), info (blue)
   - Auto-dismiss after 4 seconds, manual dismiss
   - Used for: create/delete success, save failures, cycle warnings, forbidden actions

3. **EmptyState component** — parameterized with icon, title, description, action button
   - Board empty: "No work items yet. Create your first item to get started."
   - List empty: "No items match your filters." + "Clear filters" button
   - Detail relationships empty: "No relationships yet. Add one to track dependencies."
   - Detail children empty: "No child items yet. Create one to break this work down."
   - Graph empty: "Select an item to explore its dependency graph."

4. **ErrorState component** — parameterized with error message, retry action
   - Used when API calls fail
   - Global error boundary wrapping `<App />` for catastrophic failures

5. **LoadingState component** — skeleton screens for board columns, list rows, detail page

6. **NotFoundPage** — shown for invalid routes and 404 API responses on detail page

7. **Error code → UI behavior integration**
   - Wire up the error mapping from the API contract into the mutation hooks
   - 400 → inline field error, 403 → toast, 404 → not found page, 409 → toast, 422 → inline error, 5xx → error banner with retry

### Deliverables
- Toast notifications for all mutations
- Empty states for every view
- Error handling for all API error codes
- Loading skeletons
- Settings page
- 404 page

### Tests
- Unit: Toast renders, auto-dismisses
- Unit: EmptyState renders with action button
- Unit: ErrorState renders with retry
- Unit: Error code mapping produces correct UI behavior

---

## Phase 9 — Accessibility + Performance Pass

### Goal
Ensure keyboard navigation, focus management, semantic HTML, color contrast, and performance targets are met.

### Dependencies
- All previous phases (complete UI)

### Tasks

1. **Keyboard navigation audit**
   - Tab order through sidebar, top bar, main content
   - Board: arrow keys to move between columns, Enter to open card, Escape to cancel drag
   - List: arrow keys to move between rows, Enter to expand/open
   - Detail: Tab through all editable fields
   - Modals: focus trap, Escape to close
   - Graph: Tab to select nodes, Enter to open detail panel

2. **Focus management**
   - On route change: focus main content heading
   - On modal open: focus first focusable element
   - On modal close: return focus to trigger element
   - On toast: do not steal focus (aria-live region)

3. **Semantic HTML + ARIA**
   - Board columns: `role="list"`, cards: `role="listitem"`
   - State badges: `aria-label` with full state name
   - Tags: removable tags have `aria-label="Remove tag {name}"`
   - Filter panel: `role="search"`, filter groups with `role="group"` + `aria-labelledby`
   - Drag-and-drop: aria-live announcements for drag start/drop/cancel

4. **Color contrast + non-color indicators**
   - All state badges include text labels (not color-only)
   - Blocked indicator includes icon (not just color)
   - Focus states: visible 2px ring on all interactive elements
   - Verify contrast ratio ≥ 4.5:1 for text, ≥ 3:1 for UI components

5. **Responsive and touch verification**
   - Verify all views render without horizontal scroll at 320px width (Chrome DevTools mobile emulation)
   - Verify all interactive elements (buttons, links, drag handles, filter toggles) meet 44px minimum tap target
   - Verify layout adapts at 768px tablet breakpoint (single-column reflow where needed)
   - Test board drag-and-drop on touch devices (or touch emulation)

6. **Performance optimization**
   - Route-level code splitting with `React.lazy` + `Suspense`
   - List view: virtualize rows with `@tanstack/react-virtual` if >100 items
   - Board: virtualize columns if >50 cards per column
   - Graph: limit to 100 nodes max, paginate with "Load more neighbors"
   - Measure P95 render times: board <300ms, list <300ms, detail <250ms, search <200ms

7. **Bundle analysis**
   - Run `vite-plugin-visualizer`, ensure no single chunk >200KB gzipped
   - Lazy-load graph dependencies (@xyflow/react) since it's the heaviest

### Deliverables
- All workflows navigable by keyboard
- Semantic HTML and ARIA attributes throughout
- Focus management on route changes and modals
- Performance within P95 targets
- Bundle optimized with lazy loading

### Tests
- Accessibility: axe-core automated scan (0 critical violations)
- Performance: Lighthouse performance score ≥ 90
- Unit: Focus returns to trigger after modal close

---

## Phase Summary

| Phase | Name | Key Deliverable |
|-------|------|-----------------|
| 1 | Scaffolding + API Client | App shell, routing, types, API layer |
| 2 | Board View | Three-column kanban with drag-and-drop |
| 3 | Detail + Create/Edit | Full item CRUD, inline edit, autosave |
| 4 | List + Filtering | Sortable list, filter panel, URL-persisted filters |
| 5 | Relationships | Add/remove relationships, cycle warnings |
| 6 | Graph View | Interactive dependency graph |
| 7 | Home Page | Workspace overview with quick navigation |
| 8 | Polish | Empty/error states, toasts, settings |
| 9 | Accessibility + Performance | Keyboard nav, ARIA, P95 targets, lazy loading |
