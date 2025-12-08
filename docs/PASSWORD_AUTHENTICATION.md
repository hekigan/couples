# Password Authentication Guide

## Overview

This application uses **Supabase Auth** for secure password management. Passwords are NEVER stored in the application database - they are securely hashed and stored by Supabase Auth.

## Architecture

### Two-Database System

1. **`auth.users`** (Supabase-managed)
   - Stores authentication credentials (email, password hash, etc.)
   - Managed entirely by Supabase Auth
   - Passwords are automatically hashed using bcrypt
   - NOT directly accessible from application code

2. **`public.users`** (Application-managed)
   - Stores application-specific user data (username, display name, etc.)
   - Defined in `sql/schema.sql`
   - NO password field (this is correct!)
   - Linked to `auth.users` by matching UUID

### Authentication Flow

```
┌─────────────────────────────────────────────────────────┐
│                    User Signup/Login                    │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   Supabase Auth      │
              │   (auth.users)       │
              │  - Email             │
              │  - Password Hash     │
              │  - User UUID         │
              └──────────┬───────────┘
                         │
                         │ Returns Access Token + User UUID
                         │
                         ▼
              ┌──────────────────────┐
              │  Application DB      │
              │  (public.users)      │
              │  - UUID (matches)    │
              │  - Username          │
              │  - Display Name      │
              │  - Email (copy)      │
              └──────────────────────┘
```

## Implementation Details

### Backend (Go)

**Service Layer** (`internal/services/auth_service.go:242-269`):

```go
// SignupWithPassword creates a new user in Supabase Auth
func (s *AuthService) SignupWithPassword(ctx context.Context, email, password, username string) (*OAuthSession, error) {
    // Step 1: Create user in Supabase Auth (password hashed automatically)
    resp, err := s.authClient.Signup(types.SignupRequest{
        Email:    email,
        Password: password,
        Data: map[string]interface{}{
            "username":     username,
            "display_name": username,
        },
    })
    // Step 2: Get user details from access token
    user, err := s.GetUserFromAccessToken(ctx, resp.AccessToken)
    // Step 3: Return session data (access token, refresh token, user info)
}
```

**Handler Layer** (`internal/handlers/auth.go:116-231`):

```go
// SignupPostHandler validates input and creates user
func (h *Handler) SignupPostHandler(w http.ResponseWriter, r *http.Request) {
    // Step 1: Validate form inputs (email, password, username)
    // Step 2: Call AuthService.SignupWithPassword()
    // Step 3: Create user record in public.users table
    // Step 4: Create session and redirect to home
}
```

### Frontend (HTMX)

**Signup Form** (`templates/auth/signup.html:18-57`):

```html
<form hx-post="/auth/signup"
      hx-swap="none"
      hx-disabled-elt="button[type='submit']"
      hx-indicator="#signup-loading"
      hx-on::after-request="handleSignupResponse(event)">

    <input type="text" name="username" required minlength="3" maxlength="50">
    <input type="email" name="email" required autocomplete="email">
    <input type="password" name="password" required minlength="6" autocomplete="new-password">
    <input type="password" name="password_confirm" required minlength="6">

    <button type="submit" class="secondary">
        <span id="signup-loading" class="htmx-indicator">⏳ </span>
        Create Account
    </button>
</form>
```

**Login Form** (`templates/auth/login.html:18-42`):

```html
<form hx-post="/auth/login"
      hx-swap="none"
      hx-disabled-elt="button[type='submit']"
      hx-indicator="#login-loading"
      hx-on::after-request="handleLoginResponse(event)">

    <input type="email" name="email" required autocomplete="email">
    <input type="password" name="password" required autocomplete="current-password">

    <button type="submit" class="secondary">
        <span id="login-loading" class="htmx-indicator">⏳ </span>
        Login
    </button>
</form>
```

## Security Best Practices Implemented

### Password Storage ✅
- ✅ Passwords NEVER stored in plain text
- ✅ Passwords NEVER stored in application database
- ✅ Passwords automatically hashed by Supabase Auth (bcrypt)
- ✅ Only password hashes stored in Supabase Auth's secure database

### Form Security ✅
- ✅ **HTTPS Only**: Credentials transmitted over HTTPS in production
- ✅ **Autocomplete Attributes**: Proper autocomplete for password managers
  - `autocomplete="email"` for email fields
  - `autocomplete="current-password"` for login password
  - `autocomplete="new-password"` for signup/change password
- ✅ **HTML5 Validation**: Client-side validation with `required`, `minlength`, `pattern`
- ✅ **Server-side Validation**: All inputs validated on server
- ✅ **Password Requirements**: Minimum 6 characters (enforced by Supabase Auth)
- ✅ **Password Confirmation**: Users must confirm password during signup
- ✅ **Clear on Error**: Password fields cleared after failed attempts

### HTMX Security ✅
- ✅ **Same-Origin Requests**: Forms post to same domain (CSRF protection)
- ✅ **No Password in URL**: POST requests (passwords not in query params)
- ✅ **Disabled During Submit**: `hx-disabled-elt` prevents double submission
- ✅ **Generic Error Messages**: Prevents user enumeration attacks
  - "Invalid email or password" instead of "User not found"

### Session Security ✅
- ✅ **Session Tokens**: Uses `gorilla/sessions` for secure session management
- ✅ **HTTP-Only Cookies**: Session cookies not accessible via JavaScript
- ✅ **Secure Flag**: Cookies marked secure in production (HTTPS only)
- ✅ **Access Tokens**: Supabase access tokens stored in session
- ✅ **Refresh Tokens**: Refresh tokens for token renewal

## Database Schema

**NO Password Field** (Correct!):

```sql
-- sql/schema.sql (lines 15-25)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    is_anonymous BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
-- ❌ NO password field - passwords managed by Supabase Auth
```

## Usage

### For Development

**1. Start the server:**
```bash
make dev
```

**2. Create an account:**
- Visit: http://localhost:8080/auth/signup
- Fill in:
  - Username: testuser
  - Email: test@example.com
  - Password: password123
  - Confirm Password: password123
- Click "Create Account"

**3. Login:**
- Visit: http://localhost:8080/auth/login
- Enter email and password
- Click "Login"

### For Production

**1. Create users via Supabase Dashboard:**
- Go to Supabase Dashboard → Authentication → Users
- Click "Add User" → "Create new user"
- Enter email and password
- User is automatically created in both `auth.users` and `public.users`

**2. Enable email confirmation (optional):**
- In Supabase Dashboard → Authentication → Settings
- Enable "Confirm email" to require email verification
- Update signup handler to handle email confirmation flow

## Testing

**Test Signup:**
```bash
# Using curl
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=testuser&email=test@example.com&password=password123&password_confirm=password123"

# Expected: {"success": "Account created"}
```

**Test Login:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "email=test@example.com&password=password123"

# Expected: HX-Redirect header with 200 status
```

## Common Issues

### Issue: "Authentication failed"
**Cause**: User doesn't exist in Supabase Auth
**Solution**: Create account via signup form first

### Issue: "Email already in use"
**Cause**: User already exists in Supabase Auth
**Solution**: Use login form instead, or use password recovery

### Issue: "Invalid email or password"
**Cause**: Wrong credentials
**Solution**: Check email/password, ensure caps lock is off

### Issue: "Failed to create user account"
**Cause**: User exists in Supabase Auth but not in application DB
**Solution**: Check Supabase logs, ensure `CreateOrUpdateUserFromOAuth` works correctly

## Admin User Setup

See `sql/seed.sql` (lines 4-34) for detailed instructions:

**Option 1: Signup Form (Recommended)**
1. Use signup form to create user
2. Run SQL: `UPDATE users SET is_admin = TRUE WHERE email = 'admin@example.com';`

**Option 2: Supabase Dashboard**
1. Create user in Supabase Dashboard → Authentication
2. Copy UUID
3. Update `sql/seed.sql` with UUID
4. Run: `psql -f sql/seed.sql`

**Option 3: Dev Login (Development Only)**
- Visit: http://localhost:8080/auth/dev-login-admin
- Bypasses authentication (development only!)

## Future Enhancements

Potential improvements (out of scope for current implementation):

- [ ] Password strength meter on signup form
- [ ] "Forgot password" flow (password reset)
- [ ] "Remember me" checkbox (extend session duration)
- [ ] Email verification required for signup
- [ ] Two-factor authentication (2FA/MFA)
- [ ] Rate limiting on login attempts (prevent brute force)
- [ ] Account lockout after failed attempts
- [ ] Password change functionality in profile
- [ ] Social login (Google, Facebook) - already implemented via OAuth

## Related Files

**Backend:**
- `internal/services/auth_service.go:242-269` - Signup/login methods
- `internal/handlers/auth.go:23-231` - Auth HTTP handlers
- `cmd/server/main.go:100-106` - Route definitions

**Frontend:**
- `templates/auth/signup.html` - Signup form with HTMX
- `templates/auth/login.html:18-70` - Login form with HTMX

**Database:**
- `sql/schema.sql:15-25` - Users table (no password field)
- `sql/seed.sql:4-39` - Admin user setup instructions

**Documentation:**
- `CLAUDE.md` - Project overview and guidelines
- `docs/PASSWORD_AUTHENTICATION.md` - This file
