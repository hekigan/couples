# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Couple Card Game** - a Go-based web application built with HTMX and Supabase. The game helps couples strengthen their relationship through question-based conversations. It features real-time multiplayer gameplay using Server-Sent Events (SSE), OAuth authentication, a friend system, multi-language support (EN, FR, JA), and an admin panel.

**Tech Stack:**
- Backend: Go 1.22+ with gorilla/mux router
- Database: PostgreSQL via Supabase (supabase-community/supabase-go)
- Frontend: HTMX + SASS (transitioning from vanilla JavaScript)
- Real-time: Server-Sent Events (SSE)
- Session: gorilla/sessions
- Testing: Go test + Supabase CLI (local test database)

## Build & Run Commands

```bash
# First time setup
make test-db-setup    # Setup local Supabase test database (one-time)
make deps             # Install Go and Node dependencies
make build            # Build the Go binary

# Development workflow
make run              # Build and run server (includes SASS compilation)
make dev              # Run with go run (includes SASS compilation)
make sass-watch       # Watch and auto-compile SASS in separate terminal

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
make clean            # Remove build artifacts
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
- `UserService` - User management and authentication
- `RoomService` - Game room CRUD and state management
- `GameService` - Game logic, turn management, question drawing
- `QuestionService` - Question/category management
- `AnswerService` - Answer recording
- `FriendService` - Friend requests and relationships
- `NotificationService` - User notifications
- `I18nService` - Multi-language translation system
- `RealtimeService` - SSE connection management and broadcasting
- `TemplateService` - HTML fragment rendering for HTMX (NEW)

**Service initialization in `cmd/server/main.go`:**
```go
realtimeService := services.NewRealtimeService()
userService := services.NewUserService(supabaseClient)
roomService := services.NewRoomService(supabaseClient, realtimeService)
// ... more services
handler := handlers.NewHandler(userService, roomService, ..., templateService)
```

### Handler Pattern

All HTTP handlers are methods on a single `Handler` struct (`internal/handlers/base.go`) that holds references to all services. This enables:
- Shared service access across handlers
- Easy testing via dependency injection
- Centralized template rendering

**Handler files:**
- `base.go` - Handler struct, constructor, template rendering
- `auth.go` - Authentication (login, logout, OAuth)
- `game.go` - Game endpoints (rooms, play, API)
- `room_join.go` - Join request flow
- `friends.go` - Friend system
- `notifications.go` - Notification endpoints
- `realtime.go` - SSE event streams
- `admin.go` - Admin panel

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

**Two SSE modes (transitioning from JSON to HTML):**

**Legacy JSON SSE (being phased out):**
```go
h.RoomService.GetRealtimeService().BroadcastPlayerJoined(roomID, playerData)
// Sends: event: player_joined
//        data: {"user_id":"...", "username":"..."}
```

**New HTML SSE (for HTMX):**
```go
html, _ := h.TemplateService.RenderFragment("player_joined.html", data)
h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
    Type:       "player_joined",
    Target:     "#guest-info",
    SwapMethod: "innerHTML",
    HTML:       html,
})
// Sends: event: player_joined
//        data: {"target":"#guest-info","swap":"innerHTML","html":"<div>...</div>"}
```

**HTML Templates for SSE:** Located in `templates/partials/` with corresponding data structs in `internal/services/template_service.go`.

### HTMX + Supabase Refactoring (IN PROGRESS)

**The codebase is undergoing a refactoring from vanilla JavaScript to HTMX.** This is a multi-phase project:

- ✅ **Phase 1:** Database optimization with views/indexes (COMPLETE)
- ✅ **Phase 2:** SSE HTML fragment infrastructure (COMPLETE)
- ✅ **Phase 3:** Handler integration with TemplateService (COMPLETE)
- ⏸️ **Phase 4:** Convert handlers to render HTML fragments (PENDING)
- ⏸️ **Phase 5:** HTMX integration in templates (PENDING)
- ⏸️ **Phase 6:** E2E testing (PENDING)

**See:** `PHASE1_COMPLETE.md`, `PHASE2_SSE_HTML_FRAGMENTS.md`, `PHASE3_INFRASTRUCTURE_COMPLETE.md`, `SESSION_SUMMARY.md`

**When working on handlers, prefer the new HTML fragment pattern over legacy JSON SSE.**

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

### Template System

**Two template systems coexist:**

1. **Full page templates** (`templates/*.html`): Use `layout.html` wrapper, rendered via `Handler.RenderTemplate()`
2. **HTML fragments** (`templates/partials/**/*.html`): Rendered via `TemplateService.RenderFragment()` for SSE/HTMX

**Template data structures:**
- `TemplateData` - For full pages (common fields like Title, User, Error)
- Service-specific structs - For fragments (e.g., `QuestionDrawnData`, `JoinRequestData`)

### I18n System

Multi-language support via JSON files in `static/i18n/`:
- `en.json` - English
- `fr.json` - French
- `ja.json` - Japanese

**Backend:** `I18nService` loads translations, `I18nMiddleware` sets language context
**Frontend:** JavaScript function `t(key)` for translation lookups

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
`make run` depends on `make build` and `make sass`. If you modify Go files, you must rebuild.

### 2. Use Database Views for Queries
Never manually fetch related user data. Use `rooms_with_players`, `join_requests_with_users`, or `active_games` views to avoid N+1 queries.

### 3. Service Role Key vs Anon Key
Backend uses `SUPABASE_SERVICE_ROLE_KEY` to bypass RLS. This means **all authorization must be implemented in Go handlers**. Never assume Supabase RLS protects your data.

### 4. SSE Connection Management
SSE connections are stateful. When a room is deleted, gracefully close connections. See `internal/handlers/realtime.go` for error handling pattern.

### 5. HTMX Refactoring in Progress
When modifying game/room handlers, prefer the new HTML fragment pattern (Phase 4+ work). See `PHASE3_INFRASTRUCTURE_COMPLETE.md` for usage examples.

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
├── cmd/server/main.go          # Application entry point, service init, routing
├── internal/
│   ├── handlers/               # HTTP handlers (thin layer)
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Data models (Room, User, Question, etc.)
│   └── services/               # Business logic (thick layer)
├── templates/                  # Full page HTML templates
│   ├── partials/              # HTML fragments for HTMX/SSE
│   └── layout.html            # Base layout wrapper
├── static/
│   ├── css/                   # Compiled CSS (from SASS)
│   ├── js/                    # Frontend JavaScript (being reduced)
│   └── i18n/                  # Translation JSON files
├── sass/                      # SASS source files
├── sql/                       # Database schema, views, indexes, seeds
├── scripts/                   # Shell scripts (test DB setup, etc.)
└── docs/                      # Comprehensive documentation (put all *.md files here, not at the project root)

# Generated/ignored:
├── server                     # Compiled binary (gitignored)
├── coverage.out              # Test coverage (gitignored)
└── supabase/                 # Local Supabase files (gitignored)
```

## Documentation

**Extensive documentation** exists in `docs/` and root-level markdown files. Key references:
- `README.md` - Quick start
- `docs/TESTING.md` - Comprehensive testing guide
- `docs/REALTIME_NOTIFICATIONS.md` - SSE architecture
- `SESSION_SUMMARY.md` - Recent refactoring session summary
- `PHASE*.md` - HTMX refactoring phase documentation

When making architectural changes, update relevant documentation.
