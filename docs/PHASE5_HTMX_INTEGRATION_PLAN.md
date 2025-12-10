# Phase 5: HTMX Integration - Implementation Plan

**Date:** November 10, 2025
**Status:** ✅ **COMPLETE** (Updated December 2024: Templ migration completed alongside HTMX integration)

---

## 📋 Overview

**UPDATE December 2024:** Phase 5 has been completed alongside the templ migration. HTMX is fully integrated and all templates are now type-safe templ components.

**Completed State:**
- ✅ Backend renders templ components
- ✅ Backend broadcasts HTML fragments via SSE
- ✅ Frontend uses HTMX SSE extension
- ✅ Frontend auto-updates DOM with HTMX attributes
- ✅ All 60+ templates migrated to type-safe templ components
- ✅ ~2300 lines of JavaScript SSE handling eliminated

---

## 🎯 Objectives

### 1. Add HTMX Library & SSE Extension
- Include HTMX core library (via CDN or local)
- Include HTMX SSE extension
- Configure CSP headers if needed

### 2. Convert room.html (Lobby Page)
- Replace EventSource JavaScript with HTMX SSE
- Add `hx-ext="sse"` to container
- Add `sse-connect` attribute for connection
- Add `sse-swap` attributes for each event type
- Remove ~800 lines of JavaScript

### 3. Convert play.html (Game Page)
- Replace EventSource JavaScript with HTMX SSE
- Add HTMX SSE attributes
- Remove ~1500 lines of JavaScript

### 4. Templ Migration (COMPLETED December 2024)
- ✅ All HTML templates converted to .templ files
- ✅ Type-safe component parameters
- ✅ Handlers use RenderTemplComponent() and RenderTemplFragment()
- ✅ Data structures in viewmodels and services packages
- ✅ Generated *_templ.go files committed to git

### 5. Test End-to-End (COMPLETED)
- ✅ SSE connection works
- ✅ HTML fragments render correctly
- ✅ DOM updates correctly with HTMX
- ✅ All user flows tested

---

## 📦 Step 1: Add HTMX to Project

### Option A: CDN (Recommended for Development)

Add to `templates/layout.html` before closing `</body>`:

```html
<!-- HTMX Core -->
<script src="https://unpkg.com/htmx.org@1.9.10"></script>

<!-- HTMX SSE Extension -->
<script src="https://unpkg.com/htmx.org@1.9.10/dist/ext/sse.js"></script>
```

### Option B: Local Files (Recommended for Production)

1. Download HTMX files:
```bash
cd static/js
curl -O https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js
curl -O https://unpkg.com/htmx.org@1.9.10/dist/ext/sse.js
```

2. Add to `templates/layout.html`:
```html
<script src="/static/js/htmx.min.js"></script>
<script src="/static/js/sse.js"></script>
```

### Update CSP Headers (if needed)

In `internal/middleware/cors.go`, ensure HTMX is allowed:

```go
w.Header().Set("Content-Security-Policy",
    "default-src 'self'; "+
    "script-src 'self' 'unsafe-inline' https://unpkg.com; "+ // For CDN
    "connect-src 'self';")
```

---

## 🏠 Step 2: Convert room.html

### Current Implementation (Lines 292-430)

**JavaScript EventSource:**
```javascript
// ~800 lines of SSE JavaScript
joinRequestsEventSource = new EventSource(sseUrl);

joinRequestsEventSource.addEventListener('join_request', (event) => {
    const requestData = JSON.parse(event.data);
    addJoinRequestToUI(roomId, requestData);
});

joinRequestsEventSource.addEventListener('request_accepted', (event) => {
    const data = JSON.parse(event.data);
    updateUIAfterAccept(data.guest_username);
});

// ... many more event listeners
```

### New HTMX Implementation

**Replace entire SSE JavaScript with:**

```html
{{define "content"}}
<div class="room-container"
     data-room-id="{{.Data.ID}}"
     hx-ext="sse"
     sse-connect="/api/rooms/{{.Data.ID}}/events">

    <!-- Room Header (unchanged) -->
    <div class="room-header-section">
        <!-- ... existing header content ... -->
    </div>

    <!-- Players Section -->
    <div class="room-players" id="player-list">
        <h2>Players</h2>
        <div class="player-grid">
            <!-- Owner card (static) -->
            <div class="player-card">
                <span class="player-label">Owner</span>
                <span class="player-name">{{.OwnerUsername}}</span>
            </div>

            <!-- Guest card (updates via SSE) -->
            <div class="player-card"
                 id="guest-info"
                 sse-swap="request_accepted">
                {{if .Data.GuestID}}
                <span class="player-label">Guest</span>
                <span class="player-name">{{.GuestUsername}}</span>
                {{else}}
                <span>Waiting for guest...</span>
                {{end}}
            </div>
        </div>
    </div>

    <!-- Join Requests Section (owner only) -->
    {{if .IsOwner}}
    <div id="join-requests-section">
        <h2>📬 Join Requests</h2>
        <div id="join-requests"
             sse-swap="join_request"
             hx-swap-oob="true">
            <!-- Join request cards appended here via SSE -->
        </div>
    </div>
    {{end}}

    <!-- Categories Section -->
    <div class="categories-section"
         id="categories-section"
         sse-swap="categories_updated">
        <!-- Categories grid updates via SSE -->
    </div>

    <!-- Game Started Redirect -->
    <div sse-swap="game_started" style="display:none;">
        <!-- Will receive redirect fragment -->
    </div>
</div>

<!-- Minimal JavaScript (only for non-SSE features) -->
<script>
// Keep only: copyRoomId(), category selection POST, etc.
// Remove all: EventSource, addEventListener, SSE handling
</script>
{{end}}
```

### Key Changes

**Before (JavaScript):**
```javascript
joinRequestsEventSource.addEventListener('join_request', (event) => {
    const requestData = JSON.parse(event.data);
    const html = `<div class="join-request-item">...</div>`;
    document.getElementById('join-requests').insertAdjacentHTML('beforeend', html);
});
```

**After (HTMX):**
```html
<div id="join-requests"
     sse-swap="join_request">
    <!-- HTMX automatically inserts HTML fragments here -->
</div>
```

**Lines Removed:** ~800 lines of SSE JavaScript

---

## 🎮 Step 3: Convert play.html

### Current Implementation

**JavaScript EventSource (~1500 lines):**
```javascript
const sseSource = new EventSource(`/api/rooms/${roomId}/events`);

sseSource.addEventListener('question_drawn', (event) => {
    const data = JSON.parse(event.data);
    displayQuestion(data);
});

sseSource.addEventListener('answer_submitted', (event) => {
    const data = JSON.parse(event.data);
    displayAnswer(data);
});

// ... many more handlers
```

### New HTMX Implementation

```html
{{define "content"}}
<div class="game-container"
     hx-ext="sse"
     sse-connect="/api/rooms/{{.RoomID}}/events">

    <!-- Current Question Card -->
    <div id="current-question"
         sse-swap="question_drawn">
        <!-- Question card replaced via SSE -->
        <div class="question-placeholder">
            Waiting for question...
        </div>
    </div>

    <!-- Answer Display -->
    <div id="answer-display"
         sse-swap="answer_submitted">
        <!-- Answer content replaced via SSE -->
    </div>

    <!-- Turn Indicator -->
    <div id="turn-indicator"
         sse-swap="turn_changed">
        <!-- Turn info updated via SSE -->
    </div>

    <!-- Game Finished Redirect -->
    <div sse-swap="game_finished" style="display:none;">
        <!-- Will receive redirect fragment -->
    </div>
</div>

<!-- Minimal JavaScript -->
<script>
// Keep only: form submissions (hx-post handles most)
// Remove all: EventSource, SSE handling, DOM manipulation
</script>
{{end}}
```

**Lines Removed:** ~1500 lines of SSE JavaScript

---

## 🔧 Step 4: Update HTML Fragment Templates

Some templates may need adjustments for HTMX compatibility:

### question_drawn.html

**Current (already HTMX-ready):**
```html
<div class="question-card" id="current-question">
    <div class="question-text">{{.QuestionText}}</div>
    {{if .IsMyTurn}}
    <form hx-post="/api/rooms/{{.RoomID}}/answer"
          hx-swap="outerHTML"
          hx-target="#current-question">
        <textarea name="answer" required></textarea>
        <button type="submit">Submit Answer</button>
    </form>
    {{else}}
    <div>Waiting for {{.CurrentPlayerUsername}}...</div>
    {{end}}
</div>
```

✅ **No changes needed** - already uses `hx-post` for form submission

### join_request.html

**Current (already HTMX-ready):**
```html
<div class="join-request-item" id="join-request-{{.ID}}">
    <span>{{.Username}}</span>
    <button hx-post="/api/rooms/{{.RoomID}}/join-requests/{{.ID}}/accept"
            hx-swap="outerHTML"
            hx-target="#join-request-{{.ID}}">
        Accept
    </button>
</div>
```

✅ **No changes needed** - already uses `hx-post`

---

## 🧪 Step 5: Testing Checklist

### Room Lobby (room.html)

- [ ] SSE connection establishes on page load
- [ ] Join requests appear in real-time
- [ ] Accept/reject join request works
- [ ] Guest info updates when request accepted
- [ ] Category selections sync between players
- [ ] Guest ready status updates
- [ ] Game start redirect works

### Game Play (play.html)

- [ ] Question appears when drawn
- [ ] Answer form submits correctly
- [ ] Answer display updates after submission
- [ ] Turn indicator updates
- [ ] Next question appears
- [ ] Game finished redirect works

### Error Scenarios

- [ ] SSE connection retry on disconnect
- [ ] Graceful handling of malformed HTML fragments
- [ ] Timeout handling
- [ ] Network error recovery

---

## 📊 Migration Strategy

### Option A: Big Bang (All at Once)

**Pros:**
- Clean cutover
- Immediate benefits
- Simpler to reason about

**Cons:**
- Higher risk
- Harder to debug if issues
- Requires full testing before deploy

**Steps:**
1. Add HTMX to layout.html
2. Convert room.html
3. Convert play.html
4. Test thoroughly
5. Deploy

### Option B: Gradual Migration (Recommended)

**Pros:**
- Lower risk
- Easy to rollback
- Can test incrementally

**Cons:**
- Dual-mode complexity
- Longer migration period

**Steps:**
1. Add HTMX to layout.html
2. Add feature flag in templates:
   ```go
   UseHTMX := os.Getenv("USE_HTMX") == "true"
   ```
3. Create HTMX versions: `room_htmx.html`, `play_htmx.html`
4. Test HTMX versions with subset of users
5. Monitor for issues
6. Full cutover when confident
7. Remove JavaScript versions

---

## 🚨 Potential Issues & Solutions

### Issue 1: SSE Reconnection

**Problem:** HTMX SSE extension may handle reconnection differently than vanilla EventSource

**Solution:**
```html
<div hx-ext="sse"
     sse-connect="/api/rooms/{{.RoomID}}/events"
     sse-reconnect="3000"> <!-- Reconnect after 3 seconds -->
</div>
```

### Issue 2: Multiple Swap Targets

**Problem:** One SSE event may need to update multiple DOM elements

**Solution:** Use `hx-swap-oob` (out-of-band swaps) in templates:
```html
<!-- In template -->
<div id="turn-indicator" hx-swap-oob="true">
    Turn: {{.CurrentPlayerUsername}}
</div>
<div id="question-number" hx-swap-oob="true">
    Question {{.QuestionNumber}}
</div>
```

### Issue 3: Form Submissions Still Need Some JavaScript

**Problem:** Complex forms (file uploads, validation) may still need JavaScript

**Solution:** Keep minimal JavaScript for:
- Client-side validation (before HTMX sends)
- File upload progress
- Non-SSE features (copy-to-clipboard, modals)

---

## 📝 Code to Remove

### From room.html (~800 lines)

```javascript
// DELETE ENTIRE SECTION (lines 292-1100):
let joinRequestsEventSource = null;

function connectJoinRequestsStream(roomId) {
    joinRequestsEventSource = new EventSource(sseUrl);
    joinRequestsEventSource.addEventListener('join_request', ...);
    joinRequestsEventSource.addEventListener('request_accepted', ...);
    joinRequestsEventSource.addEventListener('categories_updated', ...);
    joinRequestsEventSource.addEventListener('room_update', ...);
    joinRequestsEventSource.addEventListener('game_started', ...);
    // ... many more listeners and handlers
}

function addJoinRequestToUI(roomId, requestData) { ... }
function updateUIAfterAccept(guestUsername) { ... }
function syncCategorySelections(categoryIds) { ... }
function handleGuestReadyUpdate(isReady) { ... }
// ... many more functions
```

**REPLACE WITH:**
```html
<!-- HTMX attributes on div elements -->
```

### From play.html (~1500 lines)

```javascript
// DELETE ENTIRE SECTION:
const sseSource = new EventSource(`/api/rooms/${roomId}/events`);
sseSource.addEventListener('question_drawn', ...);
sseSource.addEventListener('answer_submitted', ...);
sseSource.addEventListener('turn_changed', ...);
sseSource.addEventListener('game_finished', ...);

function displayQuestion(data) { ... }
function displayAnswer(data) { ... }
function updateTurnIndicator(data) { ... }
// ... many more functions
```

**REPLACE WITH:**
```html
<!-- HTMX attributes on div elements -->
```

---

## ✅ Success Criteria

- ✅ HTMX library loaded and working
- ✅ SSE connections establish successfully
- ✅ HTML fragments render correctly
- ✅ All user interactions work (join, accept, play game)
- ✅ ~2300 lines of JavaScript removed
- ✅ Code is simpler and more maintainable
- ✅ Performance is equal or better
- ✅ No regressions in functionality

---

## 📈 Expected Benefits

### Code Reduction
- **Before:** 2300+ lines of JavaScript SSE handling
- **After:** ~100 lines of minimal JavaScript (utilities only)
- **Reduction:** ~95% less JavaScript code

### Maintainability
- **Before:** Complex event listeners, manual DOM manipulation, state management
- **After:** Declarative HTMX attributes, server renders all HTML
- **Benefit:** Much easier to understand and modify

### Performance
- **Before:** Client-side template rendering, JSON parsing, DOM manipulation
- **After:** Server renders HTML once, browser just inserts it
- **Benefit:** Faster updates, less client-side CPU

### Reliability
- **Before:** JavaScript errors could break SSE handling
- **After:** HTMX handles SSE robustly, with auto-retry
- **Benefit:** More reliable real-time updates

---

## 🚀 Phase 5 Status

**Infrastructure:** ✅ READY (Backend sends HTML fragments)
**Plan:** ✅ COMPLETE (This document)
**Implementation:** ⏸️ READY TO START
**Testing:** ⏸️ PENDING
**Deployment:** ⏸️ PENDING

**Estimated Effort:** 3-4 hours for implementation + testing
**Risk Level:** Medium (significant UI changes, requires thorough testing)
**Recommended Approach:** Gradual migration with feature flag

---

**Next Steps:** Follow this plan to implement HTMX integration in the frontend templates. Test thoroughly before deploying to production.
