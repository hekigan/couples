# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Couple Card Game** - a Go-based web application built with HTMX and Supabase. The game helps couples strengthen their relationship through question-based conversations. It features real-time multiplayer gameplay using Server-Sent Events (SSE), OAuth authentication, a friend system, multi-language support (EN, FR, JA), and an admin panel.

**Tech Stack:**
- Backend: Go 1.22+ with Echo framework
- Templates: templ (type-safe Go templates)
- Database: PostgreSQL via Supabase (supabase-community/supabase-go)
- Frontend: HTMX + SASS + JavaScript (bundled with esbuild)
- Bundler: esbuild (via Go API) for JavaScript concatenation and minification
- Real-time: Server-Sent Events (SSE)
- Session: gorilla/sessions
- Testing: Go test + Supabase CLI (local test database)

## Guidelines

- Respect DRY (don't repeat yourself) and KISS (keep it simple and stupid) concepts
- Keep separation of concerns (html, js, css, go)
- Clean architecture
- No JSON, use HTMX
- pages first loaded or on page refresh should use SSR. Then use HTMX for following interactions
- Use PicoCSS guidelines for CSS and HTML implementation and use semantic HTML
- Always use context7 when I need code generation, setup or configuration steps, or
library/API documentation. This means you should automatically use the Context7 MCP
tools to resolve library id and get library docs without me having to explicitly ask.

## Build & Run Commands

```bash
# First time setup
make test-db-setup    # Setup local Supabase test database (one-time)
make deps             # Install Go and Node dependencies
make build            # Build the Go binary (includes SASS + JS bundling)

# Development workflow (requires 4 terminals for full hot-reload)
make dev              # Terminal 1: Run server with Air hot-reload
make sass-watch       # Terminal 2: Watch and auto-compile SASS
make js-watch         # Terminal 3: Watch and auto-bundle JavaScript
make templ-watch      # Terminal 4: Watch and auto-generate templ components (optional, Air also runs templ generate)

# Alternative: One-time build and run
make run              # Build everything and run server (production mode)

# JavaScript bundling (esbuild)
make js-build         # Build JS bundles (production: minified + source maps)
make js-build-dev     # Build JS bundles (development: unminified)
make js-watch         # Watch and rebuild JS bundles on changes
make js-clean         # Remove generated bundles

# Templ component generation
make templ-install    # Install templ CLI (one-time setup)
make templ-generate   # Generate Go code from .templ files
make templ-watch      # Watch and regenerate templ components
make templ-clean      # Remove generated *_templ.go files

# Testing
make test             # Run unit tests (no database required)
make test-full        # Run full test suite (requires test DB running)
make test-coverage    # Run tests with coverage report
make test-db-start    # Start local Supabase test database
make test-db-stop     # Stop test database
make test-db-reset    # Reset test database (clean slate)
make test-db-status   # Check test database status

# Code quality
make fmt              # Format Go code
make lint             # Run golangci-lint
make clean            # Remove build artifacts (includes JS bundles)
```

**Running a single test:**
```bash
go test -v -run TestFunctionName ./internal/services/
```

## Architecture

### Service Layer Pattern

The application uses a **service-oriented architecture** with dependency injection. All business logic is in the service layer (`internal/services/`), handlers are thin wrappers that coordinate services.

**Key Services:**
- `SupabaseClient` - Database client (uses SERVICE_ROLE_KEY to bypass RLS)
- `BaseService` - Shared database query patterns (GetSingleRecord, GetRecords, InsertRecord, UpdateRecord, DeleteRecord, CountRecords)
- `UserService` - User management, authentication, cascade delete operations
- `RoomService` - Game room CRUD and state management
- `GameService` - Game logic coordination, turn management (delegates to other services, uses rendering.TemplService for SSE fragments)
- `QuestionService` - Question CRUD, translations, random selection, history tracking
- `CategoryService` - Category CRUD and management (split from QuestionService)
- `AnswerService` - Answer recording and retrieval
- `FriendService` - Friend requests and bidirectional relationships
- `NotificationService` - User notifications and room invitations
- `AdminService` - Admin operations (dashboard stats, bulk operations)
- `I18nService` - Multi-language translation system
- `RealtimeService` - SSE connection management and broadcasting
- `rendering.TemplService` - Minimal adapter for rendering templ components in service layer (used by GameService for SSE)

**Service initialization in `cmd/server/main.go`:**
```go
renderService := rendering.NewTemplService()  // For GameService SSE fragments
realtimeService := services.NewRealtimeService()
userService := services.NewUserService(supabaseClient)
roomService := services.NewRoomService(supabaseClient, realtimeService)
questionService := services.NewQuestionService(supabaseClient)
categoryService := services.NewCategoryService(supabaseClient)
answerService := services.NewAnswerService(supabaseClient)
gameService := services.NewGameService(supabaseClient, roomService, questionService, categoryService, answerService, realtimeService, renderService)
friendService := services.NewFriendService(supabaseClient)
i18nService := services.NewI18nService(supabaseClient, "./static/i18n")
notificationService := services.NewNotificationService(supabaseClient)
adminService := services.NewAdminService(supabaseClient)
handler := handlers.NewHandler(userService, roomService, gameService, questionService, categoryService, answerService, friendService, i18nService, notificationService, adminService)
```

### BaseService Pattern (NEW)

To reduce code duplication across services, the codebase uses a **BaseService** abstraction (`internal/services/base_service.go`) that provides common database query patterns.

**Pattern:**
```go
type MyService struct {
    *BaseService  // Embedded for method inheritance
    client *supabase.Client
}

func NewMyService(client *supabase.Client) *MyService {
    return &MyService{
        BaseService: NewBaseService(client, "MyService"),
        client:      client,
    }
}
```

**Available BaseService methods:**
- `GetSingleRecord(ctx, table, id, result)` - Fetch one record by ID
- `GetRecords(ctx, table, filters, result)` - Fetch multiple records with filters
- `GetRecordsWithLimit(ctx, table, filters, limit, offset, result)` - Paginated records
- `InsertRecord(ctx, table, data)` - Insert a new record
- `UpdateRecord(ctx, table, id, data)` - Update a record by ID
- `UpdateRecordsWithFilter(ctx, table, filters, data)` - Bulk update with filters
- `DeleteRecord(ctx, table, id)` - Delete a record by ID
- `DeleteRecordsWithFilter(ctx, table, filters)` - Bulk delete with filters
- `CountRecords(ctx, table, filters)` - Count records matching filters

**When to use custom queries instead of BaseService:**
- Complex queries requiring ORDER BY, Limit, Ilike
- Database views (rooms_with_players, active_games, etc.)
- Multi-step operations with complex business logic
- Queries requiring .Single() with multiple filters

**Services using BaseService:**
- ✅ CategoryService - 100% BaseService integration (7/7 methods)
- ✅ NotificationService - 8/9 methods use BaseService
- ✅ AnswerService - 4/7 methods use BaseService
- ✅ UserService - 8/10 methods use BaseService (including cascade delete)
- ✅ FriendService - 8/12 methods use BaseService
- ✅ AdminService - 11/13 methods use BaseService
- ❌ QuestionService - Uses BaseService where appropriate, custom queries for complex operations
- ❌ RoomService - Uses database views extensively, not refactored
- ❌ GameService - Pure coordination service, no direct database calls

### Service Helper Utilities (NEW)

**Helper files for shared service patterns:**

1. **`logging.go` / `logging_test.go`** - Standardized service logging
   - `ServiceLogger` struct for consistent logging across services
   - Log levels: Debug, Info, Error
   - Automatic service name prefixing
   - Used by BaseService for query logging

2. **`query_helpers.go` / `query_helpers_test.go`** - Complex query patterns
   - Advanced query builders for complex filters
   - Pagination helpers
   - Search query builders (Ilike, ORDER BY)
   - Used by services that need more than BaseService provides

3. **`template_models.go`** - Template data structures
   - All data structures for HTML template rendering
   - Separated from `template_service.go` for better organization
   - 356 lines of type-safe template data definitions
   - Used by TemplateService and handlers for SSE/HTMX rendering

**Pattern:**
```go
// Using ServiceLogger
logger := NewServiceLogger("MyService")
logger.Info("Operation completed", map[string]interface{}{
    "recordID": id.String(),
    "duration": elapsed,
})

// Using query helpers
query := BuildComplexQuery(client, "rooms", filters).
    OrderBy("created_at", false).
    WithPagination(page, pageSize)
```

### Handler Pattern

All HTTP handlers are methods on a single `Handler` struct (`internal/handlers/base.go`) that holds references to all services. This enables:
- Shared service access across handlers
- Easy testing via dependency injection
- Centralized template rendering

**Core handler infrastructure:**
- `base.go` - Handler struct, constructor, template rendering
- `helpers.go` - Reusable handler patterns (GetRoomFromRequest, VerifyRoomParticipant, RenderHTMLFragment)
- `helpers_test.go` - Tests for handler helper functions
- `types.go` - Common handler type definitions

**Main handler files:**
- `auth.go` - Authentication (login, logout, OAuth)
- `home.go` - Home page
- `profile.go` - User profile

**Game-related handlers:**
- `game_api.go` - Game API endpoints (start, draw, answer, finish)
- `game_handlers_test.go` - Tests for game handlers
- `ui_fragments.go` - HTMX UI fragment endpoints
- `categories.go` - Category selection and management
- `room_crud.go` - Room creation, update, deletion
- `room_display.go` - Room listing and details
- `room_join.go` - Join request flow and invitations
- `play_htmx.go` - Play page HTMX endpoints
- `game.go.old` - Legacy game handler (deprecated, split into focused files)

**Social features:**
- `friends.go` - Friend system
- `notifications.go` - Notification endpoints
- `notification_stream.go` - Notification SSE streams

**Real-time:**
- `realtime.go` - SSE event streams for game rooms

**Admin handlers (in `admin/` subdirectory):**
- `admin.go` - Admin panel main pages (outside admin/)
- `admin/admin_api.go` - Admin API handler struct and shared utilities
- `admin/admin_users.go` - User management endpoints
- `admin/admin_questions.go` - Question management endpoints
- `admin/admin_categories.go` - Category management endpoints
- `admin/admin_rooms.go` - Room management endpoints
- `admin/admin_stats.go` - Dashboard statistics endpoints
- `admin/admin_bulk.go` - Bulk operations endpoints
- `admin/api/csv.go` - CSV export utilities
- `admin/translation.go` - Translation management

**Handler Refactoring (Nov 2024):**
The monolithic `game.go` (1,700+ lines) was split into 7 focused files:
- `game_api.go` - Core game API endpoints
- `room_crud.go` - Room CRUD operations
- `room_display.go` - Room display pages
- `ui_fragments.go` - HTMX UI fragments
- `categories.go` - Category management
- `helpers.go` - Shared handler utilities
- `types.go` - Type definitions

This reduced duplication by 61% through helper extraction and improved code organization.

### Database Views for Performance

**Critical Pattern:** This codebase uses PostgreSQL **views** to eliminate N+1 query problems. Never manually join user data in Go - use the views instead.

**Available views (in `sql/views.sql`):**
- `rooms_with_players` - Room + owner/guest/current_player usernames (3 queries → 1)
- `join_requests_with_users` - Join requests + user info (11 queries → 1)
- `active_games` - Complete game state with question/category (5+ queries → 1)

**Usage pattern:**
```go
// BAD: Manual joining causes N+1 queries
room := GetRoom(roomID)
owner := GetUser(room.OwnerID)
guest := GetUser(room.GuestID)

// GOOD: Use the view
roomWithPlayers := GetRoomWithPlayers(roomID) // Single query
```

**Corresponding Go models in `internal/models/room.go`:**
- `Room` - Basic room struct
- `RoomWithPlayers` - Extends Room with player usernames
- `ActiveGame` - Complete game state (room + players + question + category)

### Real-time Architecture (SSE)

The application uses **Server-Sent Events** for real-time updates, not WebSockets.

**Flow:**
1. Client connects to SSE endpoint: `GET /api/rooms/{id}/events`
2. Server registers client in `RealtimeService.clients` map
3. When events occur, handlers call `BroadcastHTMLFragment()` or legacy JSON methods
4. `RealtimeService` sends events to all connected clients in that room
5. HTMX receives HTML fragments and swaps them into the DOM

**Do not use JSON, only HTMX**

**HTML SSE with templ components:**
```go
// In handlers - render templ component to HTML string
html, err := h.RenderTemplFragment(c, roomFragments.PlayerJoined(&data))
if err != nil {
    return err
}

// Broadcast via SSE
h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
    Type:       "player_joined",
    Target:     "#guest-info",
    SwapMethod: "innerHTML",
    HTML:       html,
})
// Sends: event: player_joined
//        data: {"target":"#guest-info","swap":"innerHTML","html":"<div>...</div>"}
```

**Templ components for SSE:** Located in `internal/views/fragments/` with type-safe Go parameters (uses `*services.DataType` or `*viewmodels.DataType`).

### Templ Migration (COMPLETED December 2024)

**The codebase has completed migration from Go's `html/template` to the templ framework.** All 60+ templates are now type-safe templ components.

**Migration highlights:**
- ✅ All HTML templates converted to `.templ` files
- ✅ Type-safe component parameters (compile-time checks)
- ✅ Full HTMX/SSE integration with templ fragments
- ✅ Handlers use `RenderTemplComponent()` for pages, `RenderTemplFragment()` for SSE
- ✅ Data structures in `viewmodels` package (GameStartedData, QuestionDrawnData, etc.)
- ✅ Hot-reload with Air + templ watch
- ✅ Old `html/template` system completely removed

**Template patterns:**
- Full pages: `h.RenderTemplComponent(c, pages.HomePage(data))`
- SSE fragments: `html, _ := h.RenderTemplFragment(c, fragments.PlayerJoined(&data))`
- GameService SSE: Uses `rendering.TemplService` adapter (breaks import cycle)

### Middleware Chain

Middleware is applied in `cmd/server/main.go` in this order:
1. `SecurityHeadersMiddleware` - Security headers (CSP, X-Frame-Options, etc.)
2. `CORSMiddleware` - CORS headers
3. `AuthMiddleware` - Session management, sets user context
4. `I18nMiddleware` - Language detection/setting
5. `AnonymousSessionMiddleware` - Creates anonymous sessions

**Route-specific middleware:**
- `RequireAuth` - Enforces authentication (redirects to login)
- `AdminPasswordGate` - Admin password gate
- `RequireAdmin` - Enforces admin role

### Database Schema Pattern

**Important:** The backend uses `SUPABASE_SERVICE_ROLE_KEY`, not the anon key. This **bypasses Row Level Security (RLS)**. Authorization is handled entirely in the Go handlers, not in the database.

**Schema files:**
- `sql/schema.sql` - Table definitions
- `sql/seed.sql` - Initial data (questions, categories, translations)
- `sql/views.sql` - Performance views
- `sql/indexes.sql` - Performance indexes

**Test Database:** Uses Supabase CLI for a local PostgreSQL instance. Schema/seed are automatically applied.

### Session & Authentication

**Session management:** Uses `gorilla/sessions` with cookie store. Session data includes:
- `user_id` - UUID of authenticated user
- `username` - Display name
- `language` - User's language preference (en, fr, ja)
- `is_anonymous` - Whether user is anonymous (guest)

**OAuth providers:** Google, Facebook, GitHub (configured via `.env`)

**Anonymous users:** Can play games but with limited features. Username is set during `/setup-username` flow.

### Template System (Templ Framework)

**All templates are type-safe templ components** organized in `internal/views/`:

**Directory structure:**
- `layouts/` - Layout wrappers (base.templ, admin.templ, header.templ, footer.templ)
- `pages/` - Full page components (home.templ, profile.templ, game/*.templ, admin/*.templ, auth/*.templ, friends/*.templ)
- `fragments/` - HTMX/SSE partial components (game/*.templ, room/*.templ, play/*.templ, admin/*.templ, etc.)

**Component patterns:**
```templ
// Page component with layout wrapper
templ HomePage(data *viewmodels.TemplateData) {
    @layouts.Base(data, HomeContent(data))
}

templ HomeContent(data *viewmodels.TemplateData) {
    <section class="home">
        <h1>Play together, learn together</h1>
        // ... content
    </section>
}

// Fragment component for HTMX/SSE
templ PlayerJoined(data *services.PlayerJoinedData) {
    <div id="guest-info">
        <p>{ data.Username } joined!</p>
    </div>
}
```

**Data structures:**
- `viewmodels.TemplateData` - For full pages (common fields: Title, User, Error, Success, Env)
- `viewmodels.*Data` - For SSE fragments used by both handlers and services (GameStartedData, QuestionDrawnData, AnswerSubmittedData)
- `services.*Data` - For most other fragments and complex data structures (30+ types organized by feature)

**Rendering methods:**
- `h.RenderTemplComponent(c, component)` - Renders full page directly to response
- `h.RenderTemplFragment(c, component)` - Renders fragment to HTML string (for SSE broadcasting)

### I18n System

Multi-language support via JSON files in `static/i18n/`:
- `en.json` - English
- `fr.json` - French
- `ja.json` - Japanese

**Backend:** `I18nService` loads translations, `I18nMiddleware` sets language context
**Frontend:** JavaScript function `t(key)` for translation lookups

### JavaScript Bundling (esbuild)

The application uses **esbuild** (via Go API) for JavaScript concatenation and minification. This provides:
- **Development mode:** Individual JS files loaded for easier debugging
- **Production mode:** Minified bundles with external source maps

**Architecture:**
```
esbuild/app-entry.js → esbuild → static/dist/app.bundle.js (61KB minified)
esbuild/admin-entry.js → esbuild → static/dist/admin.bundle.js (404B minified)
```

**Key files:**
- `esbuild/app-entry.js` - Entry point for main bundle (htmx, sse, ui-utils, modal, notifications)
- `esbuild/admin-entry.js` - Entry point for admin bundle
- `internal/build/esbuild_service.go` - Bundling service (BuildApp, BuildAdmin, Watch, Clean)
- `cmd/esbuild/main.go` - CLI tool for building bundles

**Bundle configuration:**
- **Production:** Minified + external source maps (.js.map files) + 82-94% size reduction
- **Development:** Unminified + inline source maps for debugging
- **Format:** IIFE (preserves window.* exports)
- **Target:** ES2020

**JavaScript loading (conditional):**
- Production: `<script src="/static/dist/app.bundle.js">` (internal/views/layouts/head_common.templ)
- Development: Individual files `<script src="/static/js/htmx.min.js">`, etc.
- Controlled by `Env` field in viewmodels.TemplateData (set from ENV environment variable)

**Source maps:**
- Production: External .map files (e.g., app.bundle.js.map) downloaded only when DevTools open
- Development: Inline source maps embedded in bundles
- Browser automatically maps minified code to original source via `//# sourceMappingURL` comment

**Important:** Entry point files in `esbuild/` are build-time only (not served publicly). Only bundles in `static/dist/` are served at runtime.

## Key Testing Patterns

**Test setup:** All service tests use `test_helpers.go` to create a test Supabase client that connects to the local test database (via `make test-db-start`).

**Test pattern:**
```go
func TestSomething(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    client := GetTestSupabaseClient(t)
    service := services.NewUserService(client)
    // ... test logic
}
```

**Unit tests:** Run with `make test` (uses `-short` flag, no DB)
**Integration tests:** Run with `make test-full` (requires test DB running)

## Common Gotchas

### 1. Always Build Before Running
`make run` depends on `make build`, `make sass`, `make js-build`, and `make templ-generate`. If you modify Go files, templ files, or JavaScript files, you must rebuild. In development, use `make dev` + `make sass-watch` + `make js-watch` + `make templ-watch` for full hot-reload (4 terminals).

### 2. Use Database Views for Queries
Never manually fetch related user data. Use `rooms_with_players`, `join_requests_with_users`, or `active_games` views to avoid N+1 queries.

### 3. Service Role Key vs Anon Key
Backend uses `SUPABASE_SERVICE_ROLE_KEY` to bypass RLS. This means **all authorization must be implemented in Go handlers**. Never assume Supabase RLS protects your data.

### 4. SSE Connection Management
SSE connections are stateful. When a room is deleted, gracefully close connections. See `internal/handlers/realtime.go` for error handling pattern.

### 5. Templ Components Are Type-Safe
When creating or modifying templates, remember templ components are compiled Go code. Type mismatches will fail at compile time. Use `*services.DataType` or `*viewmodels.DataType` for parameters. Run `make templ-generate` after editing `.templ` files.

### 6. UUID Parsing
Always validate UUIDs from URL params:
```go
roomID, err := uuid.Parse(vars["id"])
if err != nil {
    http.Error(w, "Invalid room ID", http.StatusBadRequest)
    return
}
```

### 7. Context Usage
Handlers receive `*http.Request` with context. Pass `r.Context()` to service methods for proper request lifecycle management.

### 8. JavaScript Bundling and ENV Variable
Templ components use the `Env` field from viewmodels.TemplateData to conditionally load bundled vs individual JavaScript files. Set `ENV=production` to load minified bundles, or `ENV=development` (or unset) to load individual files. Entry point files in `esbuild/` are build-time only - never modify files there without understanding the bundling architecture.

## Environment Setup

Copy `.env.example` to `.env` and configure:
- `SUPABASE_URL` and `SUPABASE_SERVICE_ROLE_KEY` (required)
- `SESSION_SECRET` (min 32 chars, required)
- `ADMIN_PASSWORD` (for admin panel access)
- OAuth credentials (optional, for social login)

**For testing:** Test database uses Supabase CLI defaults, no manual `.env` config needed.

## Project Structure Context

```
couple-game/
├── cmd/
│   ├── server/main.go         # Application entry point, service init, routing
│   └── esbuild/main.go        # esbuild CLI tool (build, watch, clean)
├── internal/
│   ├── build/                 # Build tools and services
│   │   └── esbuild_service.go # JavaScript bundling service
│   ├── handlers/              # HTTP handlers (thin layer)
│   │   ├── admin/            # Admin-specific handlers
│   │   │   ├── admin_api.go  # Base admin handler struct
│   │   │   ├── admin_users.go, admin_questions.go, admin_categories.go, etc.
│   │   │   └── api/csv.go    # CSV export utilities
│   │   ├── base.go           # Main Handler struct
│   │   ├── helpers.go        # Shared handler utilities
│   │   ├── types.go          # Handler type definitions
│   │   ├── game_api.go, room_crud.go, room_display.go, etc.
│   │   └── game.go.old       # Deprecated monolithic handler
│   ├── middleware/            # HTTP middleware
│   ├── models/                # Data models (Room, User, Question, etc.)
│   ├── rendering/             # Templ rendering adapters
│   │   └── templ_service.go  # Minimal adapter for service layer SSE rendering
│   ├── services/              # Business logic (thick layer)
│   │   ├── base_service.go   # Common database query patterns
│   │   ├── logging.go        # Service logging utilities
│   │   ├── query_helpers.go  # Complex query builders
│   │   ├── template_models.go # Template data structures (30+ types)
│   │   └── *_service.go      # Domain-specific services
│   ├── viewmodels/            # View data structures
│   │   └── template_data.go  # TemplateData, GameStartedData, QuestionDrawnData, etc.
│   └── views/                 # Templ components
│       ├── layouts/          # Layout wrappers (base.templ, admin.templ, header.templ, footer.templ)
│       ├── pages/            # Full page components (home.templ, profile.templ, game/*.templ, admin/*.templ, etc.)
│       │   ├── auth/        # Authentication pages
│       │   ├── game/        # Game pages
│       │   ├── admin/       # Admin pages
│       │   └── friends/     # Friends pages
│       └── fragments/        # HTMX/SSE partial components
│           ├── game/        # Game fragments
│           ├── room/        # Room fragments
│           ├── play/        # Play fragments
│           ├── admin/       # Admin fragments
│           ├── joinroom/    # Join room fragments
│           ├── friends/     # Friends fragments
│           └── notifications/ # Notification fragments
├── esbuild/                   # JavaScript bundle entry points (build-time only)
│   ├── app-entry.js          # Main app bundle entry point
│   └── admin-entry.js        # Admin bundle entry point
├── static/
│   ├── css/                  # Compiled CSS (from SASS, gitignored)
│   ├── dist/                 # Bundled JavaScript (gitignored)
│   │   ├── app.bundle.js    # Main app bundle (prod: 61KB, dev: 358KB)
│   │   ├── app.bundle.js.map # Source map (prod only)
│   │   ├── admin.bundle.js  # Admin bundle (prod: 404B, dev: 6.7KB)
│   │   └── admin.bundle.js.map # Source map (prod only)
│   ├── js/                   # Source JavaScript files (runtime)
│   │   ├── htmx.min.js      # HTMX core library
│   │   ├── sse.js           # HTMX SSE extension
│   │   ├── ui-utils.js      # Shared utilities (Toast, Loading, etc.)
│   │   ├── modal.js         # Modal system
│   │   ├── notifications.js      # Real-time notifications
│   │   └── admin.js         # Admin panel utilities
│   └── i18n/                 # Translation JSON files
├── sass/                     # SASS source files
├── sql/                      # Database schema, views, indexes, seeds
├── scripts/                  # Shell scripts (test DB setup, etc.)
└── docs/                     # Comprehensive documentation (put all *.md files here, not at the project root)

# Generated/committed:
├── internal/views/**/*_templ.go  # Generated Go code from .templ files (COMMITTED to git)

# Generated/ignored:
├── server                    # Compiled binary (gitignored)
├── coverage.out             # Test coverage (gitignored)
├── static/dist/             # JavaScript bundles (gitignored)
├── static/css/              # Compiled CSS (gitignored)
└── supabase/                # Local Supabase files (gitignored)
```

## Documentation

**Extensive documentation** exists in `docs/`. Key references:
- `README.md` - Quick start
- `docs/TESTING.md` - Comprehensive testing guide
- `docs/REALTIME_NOTIFICATIONS.md` - SSE architecture
- `docs/SESSION_SUMMARY.md` - Recent refactoring session summary
- `docs/PHASE*.md` - HTMX refactoring phase documentation

When making architectural changes, update relevant documentation.
