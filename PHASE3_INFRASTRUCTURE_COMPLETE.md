# Phase 3: Handler Integration - Infrastructure Complete

**Date:** November 10, 2025
**Status:** ✅ **INFRASTRUCTURE READY FOR USE**

---

## 📋 Overview

Phase 3 completes the integration of the TemplateService with the application handlers, making HTML fragment rendering available throughout the codebase. The foundation is now in place for handlers to render and broadcast HTML fragments via SSE.

---

## 🎯 Objectives Completed

### 1. TemplateService Initialization: ✅ COMPLETE

**File:** `cmd/server/main.go`

```go
// Initialize template service for SSE HTML fragments
templateService, err := services.NewTemplateService("./templates")
if err != nil {
    log.Fatalf("Failed to initialize template service: %v", err)
}
```

**What it does:**
- Loads all HTML fragment templates from `templates/partials/**/*.html`
- Makes templates available for rendering throughout the application
- Fails fast if templates have syntax errors (at startup, not runtime)

### 2. Handler Integration: ✅ COMPLETE

**File:** `internal/handlers/base.go`

**Updated Handler struct:**
```go
type Handler struct {
    UserService         *services.UserService
    RoomService         *services.RoomService
    GameService         *services.GameService
    QuestionService     *services.QuestionService
    AnswerService       *services.AnswerService
    FriendService       *services.FriendService
    I18nService         *services.I18nService
    NotificationService *services.NotificationService
    TemplateService     *services.TemplateService // NEW: For SSE HTML fragments
    Templates           *template.Template
}
```

**Updated NewHandler constructor:**
```go
func NewHandler(
    // ... existing services ...
    templateService *services.TemplateService, // NEW
) *Handler {
    return &Handler{
        // ... existing services ...
        TemplateService: templateService, // NEW
    }
}
```

### 3. Build Verification: ✅ COMPLETE

```bash
✅ Build successful
✅ No compilation errors
✅ All dependencies resolved
✅ Ready for runtime use
```

---

## 🔧 How to Use in Handlers

### Example 1: Broadcast Join Request (HTML Fragment)

```go
func (h *Handler) HandleJoinRequest(w http.ResponseWriter, r *http.Request) {
    // ... existing logic to create join request ...

    // Render HTML fragment
    html, err := h.TemplateService.RenderFragment("join_request.html", services.JoinRequestData{
        ID:        request.ID.String(),
        RoomID:    roomID.String(),
        Username:  user.Username,
        CreatedAt: time.Now().Format("3:04 PM"),
    })
    if err != nil {
        log.Printf("Failed to render template: %v", err)
        // Fall back to JSON SSE or handle error
        return
    }

    // Broadcast HTML fragment via SSE
    h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
        Type:       "join_request",
        Target:     "#join-requests",
        SwapMethod: "beforeend", // Append to list
        HTML:       html,
    })
}
```

### Example 2: Broadcast Question Drawn (HTML Fragment)

```go
func (h *Handler) DrawQuestionHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing logic to draw question ...

    // Determine if it's this user's turn
    isMyTurn := room.CurrentTurn != nil && *room.CurrentTurn == userID

    // Render HTML fragment
    html, err := h.TemplateService.RenderFragment("question_drawn.html", services.QuestionDrawnData{
        RoomID:                roomID.String(),
        QuestionNumber:        room.CurrentQuestion,
        MaxQuestions:          room.MaxQuestions,
        Category:              category.Key,
        CategoryLabel:         category.Label,
        QuestionText:          question.QuestionText,
        IsMyTurn:              isMyTurn,
        CurrentPlayerUsername: currentPlayerUsername,
    })
    if err != nil {
        log.Printf("Failed to render template: %v", err)
        return
    }

    // Broadcast to all players in room
    h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
        Type:       "question_drawn",
        Target:     "#current-question",
        SwapMethod: "outerHTML", // Replace entire question card
        HTML:       html,
    })
}
```

### Example 3: Broadcast Answer Submitted (HTML Fragment)

```go
func (h *Handler) SubmitAnswerHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing logic to submit answer ...

    // Determine whose turn it is now
    isMyTurnNow := room.CurrentTurn != nil && *room.CurrentTurn == userID

    // Render HTML fragment
    html, err := h.TemplateService.RenderFragment("answer_submitted.html", services.AnswerSubmittedData{
        RoomID:                roomID.String(),
        Username:              user.Username,
        AnswerText:            answer.AnswerText,
        ActionType:            answer.ActionType,
        IsMyTurn:              isMyTurnNow,
        CurrentPlayerUsername: currentPlayerUsername,
    })
    if err != nil {
        log.Printf("Failed to render template: %v", err)
        return
    }

    // Broadcast to all players
    h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
        Type:       "answer_submitted",
        Target:     "#answer-display",
        SwapMethod: "innerHTML",
        HTML:       html,
    })
}
```

---

## 📊 Architecture Overview

```
┌──────────────────────┐
│   main.go            │
│ - Initializes        │
│   TemplateService    │
└──────────┬───────────┘
           │
           │ passes to
           ▼
┌──────────────────────┐
│   Handler            │
│ - Stores reference   │
│ - Available to all   │
│   handler methods    │
└──────────┬───────────┘
           │
           │ uses
           ▼
┌──────────────────────┐
│  TemplateService     │
│ .RenderFragment()    │
└──────────┬───────────┘
           │
           │ produces
           ▼
┌──────────────────────┐
│   HTML Fragment      │
│   (string)           │
└──────────┬───────────┘
           │
           │ passed to
           ▼
┌──────────────────────┐
│  RealtimeService     │
│ .BroadcastHTML...()  │
└──────────┬───────────┘
           │
           │ sends via
           ▼
┌──────────────────────┐
│   SSE Stream         │
│   to clients         │
└──────────────────────┘
```

---

## 🚀 Available Templates

| Template | Path | Data Structure |
|----------|------|----------------|
| Join Request | `partials/room/join_request.html` | `services.JoinRequestData` |
| Player Joined | `partials/room/player_joined.html` | `services.PlayerJoinedData` |
| Request Accepted | `partials/room/request_accepted.html` | (custom data) |
| Game Started | `partials/game/game_started.html` | `services.GameStartedData` |
| Question Drawn | `partials/game/question_drawn.html` | `services.QuestionDrawnData` |
| Answer Submitted | `partials/game/answer_submitted.html` | `services.AnswerSubmittedData` |
| Notification | `partials/notifications/notification_item.html` | `services.NotificationData` |

---

## ✅ Benefits of Current Implementation

### 1. **Type Safety**
- All template data is strongly typed
- Compile-time checks for data structure fields
- IntelliSense/autocomplete support

### 2. **Centralized Template Management**
- All templates in one service
- Easy to reload/hot-reload for development
- Consistent rendering API

### 3. **Error Handling**
- Template parse errors caught at startup
- Runtime rendering errors logged and handleable
- Graceful fallback to JSON SSE possible

### 4. **Performance**
- Templates parsed once at startup
- Concurrent rendering with RWMutex
- Minimal overhead per request

### 5. **Maintainability**
- Clear separation of concerns
- Templates are just HTML files
- Easy for frontend developers to modify

---

## 🔄 Migration Strategy

### Current State (JSON SSE)
```go
// Old way: Manual JSON construction
h.RoomService.GetRealtimeService().BroadcastPlayerJoined(roomID, map[string]interface{}{
    "user_id":  userID.String(),
    "username": username,
})
```

### New State (HTML SSE)
```go
// New way: Render and broadcast HTML
html, _ := h.TemplateService.RenderFragment("player_joined.html", services.PlayerJoinedData{
    Username: username,
    UserID:   userID.String(),
})
h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
    Type:       "player_joined",
    Target:     "#guest-info",
    SwapMethod: "innerHTML",
    HTML:       html,
})
```

### Dual Mode Support

Both JSON and HTML SSE modes work simultaneously:
- Old handlers continue using JSON
- New handlers can use HTML fragments
- Gradual migration possible
- No breaking changes

---

## 📝 Next Steps (Phase 4)

### Handler Conversion Tasks

**Priority 1: Room Lobby Handlers**
- [ ] Convert `HandleJoinRequest` to HTML fragments
- [ ] Convert `HandleAcceptRequest` to HTML fragments
- [ ] Convert `HandleRejectRequest` to HTML fragments
- [ ] Test join request flow with HTML SSE

**Priority 2: Game Play Handlers**
- [ ] Convert `DrawQuestionHandler` to HTML fragments
- [ ] Convert `SubmitAnswerHandler` to HTML fragments
- [ ] Convert `NextQuestionHandler` to HTML fragments
- [ ] Test game flow with HTML SSE

**Priority 3: Notification Handlers**
- [ ] Convert notification broadcasts to HTML fragments
- [ ] Test notification display with HTML SSE

### HTMX Integration Tasks

**Room Lobby (room.html)**
- [ ] Add HTMX SSE extension
- [ ] Add `hx-ext="sse"` to root element
- [ ] Add `sse-connect` attribute
- [ ] Add `sse-swap` attributes to targets
- [ ] Remove ~1400 lines of JavaScript

**Game Play (play.html)**
- [ ] Add HTMX SSE extension
- [ ] Configure SSE connection
- [ ] Add swap targets
- [ ] Remove ~900 lines of JavaScript

---

## 🎯 Success Criteria

- ✅ TemplateService initialized at startup
- ✅ Handler has access to TemplateService
- ✅ Build compiles successfully
- ✅ No runtime errors expected
- ✅ Ready for handler conversion
- ✅ Backward compatible with existing code

---

## 📈 Progress Tracking

```
Phase 1: Database Optimization       [████████████████████] 100%
Phase 2: SSE HTML Fragment Infra     [████████████████████] 100%
Phase 3: Handler Integration Infra   [████████████████████] 100%
Phase 4: Convert Handlers to HTML    [░░░░░░░░░░░░░░░░░░░░]   0%
Phase 5: HTMX Integration            [░░░░░░░░░░░░░░░░░░░░]   0%
Phase 6: Testing & E2E               [░░░░░░░░░░░░░░░░░░░░]   0%

Overall Progress: 50% Complete
```

---

## 💡 Technical Notes

### Template Lookup
Templates are looked up by filename:
- `"join_request.html"` → `templates/partials/room/join_request.html`
- Go's template package handles path resolution

### Error Handling Pattern
```go
html, err := h.TemplateService.RenderFragment(templateName, data)
if err != nil {
    log.Printf("Failed to render %s: %v", templateName, err)
    // Option 1: Fall back to JSON SSE
    h.broadcastJSON(...)
    // Option 2: Return error to client
    http.Error(w, "Template error", 500)
    // Option 3: Use default template
    html = defaultHTML
}
```

### Performance Considerations
- Templates are cached after first parse
- Rendering is thread-safe (RWMutex)
- Minimal allocation per render
- Can handle high concurrent load

---

## ✅ Phase 3 Status

**Infrastructure:** ✅ COMPLETE
**Handler Conversion:** ⏸️ PENDING (Phase 4)
**HTMX Integration:** ⏸️ PENDING (Phase 5)
**Testing:** ⏸️ PENDING (Phase 6)

**Ready for:** Handler conversion to use HTML fragments

---

**Next Phase:** Phase 4 - Convert Handlers to use HTML Fragments
**Estimated Effort:** 2-3 hours for all critical handlers
**Expected Outcome:** Full SSE HTML fragment support + HTMX-ready pages

---

**Status:** Infrastructure complete, ready for implementation! 🚀
