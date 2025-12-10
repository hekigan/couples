# API Documentation - v1

## Overview

This application uses a clean, versioned API structure with all endpoints under `/api/v1/*` for the main API and `/admin/api/v1/*` for admin operations. This ensures a stable API contract and allows for future evolution without breaking changes.

## API Structure

```
Public Routes (Unversioned - UI Pages)
├── /                          # Home page
├── /health                    # Health check
├── /auth/*                    # Authentication pages
├── /profile/*                 # Profile pages
├── /friends/*                 # Friends pages
└── /game/*                    # Game pages

API v1 (Current Version)
├── /api/v1/rooms/*            # Room management
├── /api/v1/categories/*       # Categories
├── /api/v1/friends/*          # Friends API
├── /api/v1/join-requests/*    # Join requests
├── /api/v1/invitations/*      # Room invitations
├── /api/v1/notifications/*    # Notifications
└── /api/v1/stream/*           # Real-time SSE streams

Admin API v1 (Current Version)
├── /admin/*                   # Admin UI pages (unversioned)
├── /admin/api/v1/users/*      # User management
├── /admin/api/v1/questions/*  # Question management
├── /admin/api/v1/categories/* # Category management
├── /admin/api/v1/rooms/*      # Room management
├── /admin/api/v1/dashboard/*  # Dashboard stats
├── /admin/api/v1/csv/*        # CSV operations
└── /admin/api/v1/translations/* # Translation management
```

## Middleware & Security

All API endpoints are protected with:

- **Authentication**: Session-based auth (gorilla/sessions + Supabase)
- **Rate Limiting**: 20 req/sec, burst 50 (except SSE endpoints)
- **CSRF Protection**: Cookie-based tokens (except GET/HEAD/OPTIONS and SSE)
- **CORS**: Configurable via environment variables
- **Security Headers**: CSP, X-Frame-Options, X-Content-Type-Options, etc.
- **Gzip Compression**: Automatic response compression

## API v1 Endpoints

### Room Management (`/api/v1/rooms`)

**Game State Management:**
| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| DELETE | `/:id` | Delete a room | ✅ | ✅ |
| POST | `/:id/leave` | Leave a room | ✅ | ✅ |
| POST | `/:id/start` | Start game | ✅ | ✅ |
| POST | `/:id/guest-ready` | Mark guest as ready | ✅ | ✅ |
| POST | `/:id/typing` | Typing indicator | ✅ | ✅ |
| POST | `/:id/draw` | Draw next question | ✅ | ✅ |
| POST | `/:id/answer` | Submit answer | ✅ | ✅ |
| POST | `/:id/finish` | Finish game | ✅ | ✅ |
| POST | `/:id/next-question` | Get next question (HTMX) | ✅ | ✅ |

**UI Fragments (HTMX):**
| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `/:id/start-button` | Start button state | ✅ | ✅ |
| GET | `/:id/ready-button` | Ready button state | ✅ | ✅ |
| GET | `/:id/status-badge` | Status badge | ✅ | ✅ |
| GET | `/:id/turn-indicator` | Turn indicator | ✅ | ✅ |
| GET | `/:id/question-card` | Question card | ✅ | ✅ |
| GET | `/:id/game-forms` | Game forms | ✅ | ✅ |
| GET | `/:id/game-content` | Game content | ✅ | ✅ |
| GET | `/:id/progress-counter` | Progress counter | ✅ | ✅ |

**Categories:**
| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `/:id/categories` | Get room categories | ✅ | ✅ |
| POST | `/:id/categories` | Update categories | ✅ | ✅ |
| POST | `/:id/categories/toggle` | Toggle category | ✅ | ✅ |

**Join Requests:**
| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `/:id/join-requests` | List join requests | ✅ | ✅ |
| GET | `/:id/join-requests-json` | Get as JSON | ✅ | ✅ |
| GET | `/:id/join-requests-count` | Get count | ✅ | ✅ |
| GET | `/:id/my-join-request` | Check my request | ✅ | ✅ |
| POST | `/:id/cancel-my-request` | Cancel my request | ✅ | ✅ |

### Categories (`/api/v1/categories`)

| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `` | List all categories | ✅ | ✅ |

### Friends (`/api/v1/friends`)

| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `/list` | Get friends (JSON) | ✅ | ✅ |
| GET | `/list-html` | Get friends (HTML) | ✅ | ✅ |

### Join Requests (`/api/v1/join-requests`)

| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| POST | `` | Create join request | ✅ | ✅ |
| GET | `/my-requests` | My join requests | ✅ | ✅ |
| GET | `/my-accepted` | My accepted requests | ✅ | ✅ |
| POST | `/:request_id/accept` | Accept request | ✅ | ✅ |
| POST | `/:request_id/reject` | Reject request | ✅ | ✅ |

### Invitations (`/api/v1/invitations`)

| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| POST | `` | Send invitation | ✅ | ✅ |
| DELETE | `/:room_id/:invitee_id` | Cancel invitation | ✅ | ✅ |

### Notifications (`/api/v1/notifications`)

| Method | Endpoint | Description | Auth | CSRF |
|--------|----------|-------------|------|------|
| GET | `` | List notifications | ✅ | ✅ |
| GET | `/unread-count` | Unread count | ✅ | ✅ |
| POST | `/:id/read` | Mark as read | ✅ | ✅ |
| POST | `/read-all` | Mark all read | ✅ | ✅ |

### Real-time Streams (`/api/v1/stream`)

**Server-Sent Events (SSE) - No rate limiting, No CSRF:**

| Method | Endpoint | Description | Auth | Rate Limit | CSRF |
|--------|----------|-------------|------|------------|------|
| GET | `/rooms/:id/events` | Room events (SSE) | ✅ | ❌ | ❌ |
| GET | `/rooms/:id/players` | Player updates (SSE) | ✅ | ❌ | ❌ |
| GET | `/rooms/:id/state` | Room state (SSE) | ✅ | ❌ | ❌ |
| GET | `/user/events` | User notifications (SSE) | ✅ | ❌ | ❌ |
| GET | `/notifications` | Notification stream (SSE) | ✅ | ❌ | ❌ |

## Admin API v1 Endpoints

### Users Management (`/admin/api/v1/users`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/list` | List users (paginated) |
| GET | `/new` | Create form |
| POST | `` | Create user |
| GET | `/:id/edit-form` | Edit form |
| PUT | `/:id` | Update user |
| POST | `/:id/toggle-admin` | Toggle admin |
| DELETE | `/:id` | Delete user |
| POST | `/bulk-delete` | Bulk delete |

### Questions Management (`/admin/api/v1/questions`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/list` | List questions |
| GET | `/new` | Create form |
| POST | `` | Create question |
| GET | `/:id/edit-form` | Edit form |
| PUT | `/:id` | Update question |
| DELETE | `/:id` | Delete question |
| POST | `/bulk-delete` | Bulk delete |

### Categories Management (`/admin/api/v1/categories`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/list` | List categories |
| GET | `/new` | Create form |
| POST | `` | Create category |
| GET | `/:id/edit-form` | Edit form |
| PUT | `/:id` | Update category |
| DELETE | `/:id` | Delete category |
| POST | `/bulk-delete` | Bulk delete |

### Rooms Management (`/admin/api/v1/rooms`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/list` | List rooms |
| GET | `/:id/details` | Room details |
| POST | `/:id/close` | Close room |
| POST | `/bulk-close` | Bulk close |

### Dashboard (`/admin/api/v1/dashboard`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/stats` | Dashboard stats |

### CSV Operations (`/admin/api/v1/csv`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/questions/export` | Export questions CSV |
| POST | `/questions/import` | Import questions CSV |
| GET | `/questions/template` | CSV template |
| GET | `/categories/export` | Export categories CSV |

### Translations (`/admin/api/v1/translations`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/languages` | List languages |
| GET | `/:lang_code` | Get translations |
| PUT | `/:lang_code` | Update translation |
| POST | `` | Create translation |
| DELETE | `/:lang_code` | Delete translation |
| GET | `/:lang_code/export` | Export translations |
| POST | `/:lang_code/import` | Import translations |
| GET | `/validate` | Validate keys |
| POST | `/language/add` | Add language |

## Code Examples

### JavaScript Fetch

```javascript
// Room API call with authentication
fetch('/api/v1/rooms/123/start', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': getCsrfToken()
  }
})
.then(response => response.json())
.then(data => console.log(data));
```

### HTMX

```html
<!-- HTMX button with v1 API -->
<button hx-post="/api/v1/rooms/123/start"
        hx-swap="outerHTML"
        hx-target="#start-button">
  Start Game
</button>

<!-- HTMX polling with v1 API -->
<div hx-get="/api/v1/rooms/123/status-badge"
     hx-trigger="every 2s"
     hx-swap="outerHTML">
</div>
```

### Server-Sent Events (SSE)

```javascript
// Connect to SSE endpoint
const eventSource = new EventSource('/api/v1/stream/rooms/123/events');

// Listen for room events
eventSource.addEventListener('player_joined', (event) => {
  const data = JSON.parse(event.data);
  htmx.swap(data.target, data.html, {swapStyle: data.swap});
});

// Handle connection errors
eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  eventSource.close();
};
```

## Development Tools

### Route Registry

In development mode (`ENV=development`), the server prints route statistics on startup:

```
================================================================================
ROUTE STATISTICS
================================================================================
Total Routes:         208
API v1 Routes:        73
Admin API v1 Routes:  87
Unversioned Routes:   48 (UI pages)
================================================================================
```

### Full Route Inventory

To see a complete list of all registered routes, uncomment in `cmd/server/main.go`:

```go
// PrintRouteRegistry(e)
```

This prints a detailed inventory grouped by API version.

## Best Practices

### API Development
1. **Always use versioned endpoints** (`/api/v1/*`)
2. **Follow REST conventions** for resource naming
3. **Use proper HTTP methods** (GET, POST, PUT, DELETE)
4. **Return appropriate status codes** (200, 201, 400, 404, 500, etc.)
5. **Document new endpoints** in this file

### Frontend Development
1. **Use HTMX for UI interactions** when possible
2. **Use SSE for real-time updates** (rooms, notifications)
3. **Always include CSRF tokens** for POST/PUT/DELETE requests
4. **Handle errors gracefully** with user-friendly messages

### Security
1. **Never bypass authentication** on protected endpoints
2. **Always validate user input** on the server
3. **Use parameterized queries** to prevent SQL injection
4. **Sanitize HTML output** to prevent XSS
5. **Rate limit sensitive operations** (login, signup, etc.)

## Error Handling

All API endpoints return consistent error responses:

### JSON Errors (API requests)
```json
{
  "error": "Invalid room ID"
}
```

### HTML Errors (HTMX requests)
```html
<div class='error'>Invalid room ID</div>
```

### Common Status Codes
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Not authorized (admin only)
- `404 Not Found` - Resource doesn't exist
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Future Versioning

When breaking changes are needed, create a new version:

1. Create `routes_api_v2.go` and `routes_admin_v2.go`
2. Register new routes under `/api/v2/*` and `/admin/api/v2/*`
3. Keep v1 routes active for backward compatibility
4. Update documentation with migration guide
5. Eventually deprecate v1 after migration period

---

**Current Version**: v1
**Last Updated**: 2025-12-10
**Total Routes**: 208 (73 API + 87 Admin + 48 UI)
