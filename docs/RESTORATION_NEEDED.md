# ⚠️ Files That Need Restoration

## Overview

Some implementation files were accidentally deleted and need to be restored. Below is the list of files and their purpose.

## 🔴 Critical Files to Restore

### 1. OAuth Implementation Files

#### `internal/services/auth_service.go` (240 lines)
**Purpose**: OAuth authentication service with Supabase GoTrue  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- `AuthService` struct
- OAuth provider support (Google, Facebook, GitHub)
- Token management functions
- User information extraction
- Session handling

**Key Functions**:
```go
func NewAuthService(supabase *SupabaseClient) (*AuthService, error)
func (s *AuthService) GetOAuthURL(provider OAuthProvider) (string, error)
func (s *AuthService) GetUserFromAccessToken(ctx context.Context, accessToken string) (*OAuthUser, error)
func (s *AuthService) CreateOrUpdateUserFromOAuth(ctx context.Context, oauthUser *OAuthUser) (*models.User, error)
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*OAuthSession, error)
func (s *AuthService) SignOut(ctx context.Context, accessToken string) error
```

#### `templates/auth/oauth-callback.html` (45 lines)
**Purpose**: OAuth callback page that extracts tokens from URL fragment  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- JavaScript to extract tokens from URL fragment
- Auto-submit form to server
- Loading indicator
- Error handling

### 2. Friend System Files

#### `internal/handlers/friend.go` (275 lines)
**Purpose**: HTTP handlers for friend management  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- `FriendListHandler` - Display friends and invitations
- `AddFriendHandler` - Add friend form and processing
- `AcceptFriendHandler` - Accept invitations
- `DeclineFriendHandler` - Decline invitations
- `RemoveFriendHandler` - Remove friends

#### `templates/friends/add.html` (180 lines)
**Purpose**: Add friend page with search and User ID sharing  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- Search form for email/User ID
- User ID display with copy button
- Help section
- Success/error messages

#### `templates/friends/list.html` (190 lines)
**Purpose**: Friends list with pending invitations  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- Pending invitations section
- Friends grid with cards
- Accept/decline buttons
- Play and remove buttons
- Empty state

### 3. Test Files

#### `internal/middleware/auth_test.go` (150 lines)
**Purpose**: Tests for authentication middleware  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- `TestGetUserIDFromContext`
- `TestIsAnonymousUser`
- `TestIsAdminUser`
- `TestIsAuthenticated`
- `TestGetUserFromContext`

#### `internal/models/user_test.go` (120 lines)
**Purpose**: Tests for user models  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
- `TestUser_IsAnonymous`
- `TestUser_Validation`

#### `internal/models/errors.go` (30 lines)
**Purpose**: Common error definitions  
**Status**: ❌ DELETED - NEEDS RESTORATION

**Contains**:
```go
var (
    ErrInvalidUserName
    ErrEmailRequired
    ErrRoomFull
    ErrRoomNotFound
    ErrUserNotFound
    ErrUnauthorized
    ErrInvalidRoomID
    ErrInvalidUserID
    ErrGameNotStarted
    ErrGameAlreadyEnded
    ErrNotYourTurn
)
```

## ✅ How to Restore

### Option 1: From Git (Recommended)
If you have git history:
```bash
# Check git log for the deleted files
git log --all --full-history -- internal/services/auth_service.go

# Restore from a specific commit
git checkout <commit-hash> -- internal/services/auth_service.go
git checkout <commit-hash> -- internal/handlers/friend.go
# ... repeat for all files
```

### Option 2: Manual Recreation
The files were created during our November 6 session. Key details:

1. **auth_service.go**: Uses `gotrue-go` client, implements OAuth for 3 providers
2. **friend.go**: CRUD handlers with HTMX support, proper authorization
3. **Templates**: Use HTMX for dynamic updates, mobile-responsive
4. **Tests**: Basic unit tests using Go testing framework

### Option 3: Reference Implementation
Check these resources for implementation details:
- OAuth: See `docs/OAUTH_GUIDE.md` for architecture
- Friends: See `docs/FRIEND_SYSTEM_GUIDE.md` for details
- The handlers in `internal/handlers/auth.go` show OAuth integration patterns

## 🔧 Build Status After Restoration

Once all files are restored, verify with:

```bash
# Should compile successfully
go build -o server ./cmd/server/main.go

# Should pass
go test ./...
```

## 📋 Checklist

- [ ] `internal/services/auth_service.go` restored
- [ ] `internal/handlers/friend.go` restored
- [ ] `templates/auth/oauth-callback.html` restored
- [ ] `templates/friends/add.html` restored
- [ ] `templates/friends/list.html` restored
- [ ] `internal/middleware/auth_test.go` restored
- [ ] `internal/models/user_test.go` restored
- [ ] `internal/models/errors.go` restored
- [ ] Code compiles successfully
- [ ] Tests pass

## ⚠️ Important Notes

### Files That Must Work Together

1. **OAuth Flow**:
   - `auth_service.go` ← Core OAuth logic
   - `auth.go` handlers ← Uses auth service
   - `oauth-callback.html` ← Token extraction
   - `login.html` ← OAuth buttons

2. **Friend System**:
   - `friend_service.go` ← Already exists (backend)
   - `friend.go` handlers ← Needs restoration (HTTP layer)
   - `list.html` & `add.html` ← Needs restoration (UI)

3. **Routes**:
   The routes in `cmd/server/main.go` reference these handlers:
   ```go
   // OAuth routes - need auth_service.go
   authRouter.HandleFunc("/oauth/google", h.OAuthGoogleHandler)
   authRouter.HandleFunc("/oauth/facebook", h.OAuthFacebookHandler)
   authRouter.HandleFunc("/oauth/github", h.OAuthGithubHandler)
   authRouter.HandleFunc("/oauth/callback", h.OAuthCallbackHandler)
   authRouter.HandleFunc("/oauth/token", h.OAuthTokenHandler)
   
   // Friend routes - need friend.go
   friendRouter.HandleFunc("", h.FriendListHandler)
   friendRouter.HandleFunc("/add", h.AddFriendHandler)
   friendRouter.HandleFunc("/{id}/accept", h.AcceptFriendHandler)
   friendRouter.HandleFunc("/{id}/decline", h.DeclineFriendHandler)
   friendRouter.HandleFunc("/{id}", h.RemoveFriendHandler)
   ```

## 💡 Quick Reference

### Dependencies
```go
// auth_service.go imports
"github.com/supabase-community/gotrue-go"

// friend.go imports  
"github.com/gorilla/mux"
```

### Route Setup
All routes are already configured in `main.go`. Just need to restore the handler implementations.

---

**Status**: Restoration in progress  
**Priority**: HIGH - These files are needed for OAuth and Friend features  
**Impact**: Without these files, OAuth and Friend features won't work

**Next Steps**: Restore files using one of the methods above, then test build and functionality.

