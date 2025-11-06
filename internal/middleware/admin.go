package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// AdminPasswordGate checks for admin password authentication
func AdminPasswordGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if admin password is set in environment
		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminPassword == "" {
			// If no admin password is set, deny access for security
			http.Error(w, "Admin access not configured", http.StatusForbidden)
			return
		}

		// Get session
		session, err := Store.Get(r, "couple-card-game-session")
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Check if admin is already authenticated in session
		adminAuth, ok := session.Values["admin_authenticated"].(bool)
		if ok && adminAuth {
			// Already authenticated, proceed
			next.ServeHTTP(w, r)
			return
		}

		// Check for password in request (form or query)
		password := r.FormValue("admin_password")
		if password == "" {
			password = r.URL.Query().Get("admin_password")
		}

		// If password provided, verify it
		if password != "" {
			if password == adminPassword {
				// Correct password, save to session
				session.Values["admin_authenticated"] = true
				if err := session.Save(r, w); err != nil {
					http.Error(w, "Failed to save session", http.StatusInternalServerError)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			// Wrong password
			http.Error(w, "Invalid admin password", http.StatusUnauthorized)
			return
		}

		// No authentication, show password prompt
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Admin Access</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .auth-box {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.3);
            max-width: 400px;
            width: 100%;
        }
        h1 {
            margin: 0 0 20px 0;
            color: #333;
            font-size: 24px;
        }
        p {
            color: #666;
            margin-bottom: 20px;
        }
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 16px;
            box-sizing: border-box;
            margin-bottom: 20px;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 12px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: background 0.3s;
        }
        button:hover {
            background: #5568d3;
        }
    </style>
</head>
<body>
    <div class="auth-box">
        <h1>🔒 Admin Access</h1>
        <p>Please enter the admin password to continue.</p>
        <form method="POST">
            <input type="password" name="admin_password" placeholder="Admin Password" required autofocus>
            <button type="submit">Access Admin Panel</button>
        </form>
    </div>
</body>
</html>
		`))
	})
}

// RequireAdmin ensures user has admin privileges (checks user's is_admin flag)
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by AuthMiddleware)
		userIDVal := r.Context().Value(UserIDKey)
		if userIDVal == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, ok := userIDVal.(uuid.UUID)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Get session to check admin flag
		session, err := Store.Get(r, "couple-card-game-session")
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Check if user is admin (stored in session from login)
		isAdmin, ok := session.Values["is_admin"].(bool)
		if !ok || !isAdmin {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// User is admin, set in context for handlers to use if needed
		ctx := context.WithValue(r.Context(), "is_admin", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LogoutAdminHandler logs out admin session
func LogoutAdminHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "couple-card-game-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Clear admin authentication
	delete(session.Values, "admin_authenticated")
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
