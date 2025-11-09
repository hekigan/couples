# 🔐 OAuth Setup Guide

Configure OAuth authentication for Google, Facebook, and GitHub.

## Google OAuth Setup

### 1. Create Google OAuth Application

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Navigate to **APIs & Services** > **Credentials**
4. Click **Create Credentials** > **OAuth 2.0 Client ID**
5. Configure consent screen:
   - Application name: "Couple Card Game"
   - User support email: your email
   - Developer contact: your email
6. Create OAuth client:
   - Application type: **Web application**
   - Authorized redirect URIs:
     - `http://localhost:8080/auth/oauth/callback` (dev)
     - `https://yourdomain.com/auth/oauth/callback` (prod)

### 2. Configure Environment

Add to `.env`:

```env
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/oauth/callback
```

### 3. Test

1. Start server: `./server`
2. Go to login page
3. Click "Continue with Google"
4. Authorize the application
5. You should be redirected and logged in

## Facebook OAuth Setup

### 1. Create Facebook App

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Click **My Apps** > **Create App**
3. Choose **Consumer** app type
4. Fill in app details:
   - App name: "Couple Card Game"
   - Contact email: your email
5. Go to **Settings** > **Basic**
6. Note your **App ID** and **App Secret**
7. Add **Facebook Login** product
8. Configure OAuth redirect URIs:
   - Valid OAuth Redirect URIs: `http://localhost:8080/auth/oauth/callback`

### 2. Configure Environment

Add to `.env`:

```env
FACEBOOK_APP_ID=your-app-id
FACEBOOK_APP_SECRET=your-app-secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/auth/oauth/callback
```

### 3. Test

1. Restart server
2. Click "Continue with Facebook"
3. Authorize the application

## GitHub OAuth Setup

### 1. Create GitHub OAuth App

1. Go to [GitHub Settings](https://github.com/settings/developers)
2. Click **OAuth Apps** > **New OAuth App**
3. Fill in details:
   - Application name: "Couple Card Game"
   - Homepage URL: `http://localhost:8080`
   - Authorization callback URL: `http://localhost:8080/auth/oauth/callback`
4. Click **Register application**
5. Generate a new client secret
6. Note your **Client ID** and **Client Secret**

### 2. Configure Environment

Add to `.env`:

```env
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
GITHUB_REDIRECT_URL=http://localhost:8080/auth/oauth/callback
```

### 3. Test

1. Restart server
2. Click "Continue with GitHub"
3. Authorize the application

## Production Deployment

### Update Redirect URLs

For production, update all redirect URLs to your domain:

```env
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/oauth/callback
FACEBOOK_REDIRECT_URL=https://yourdomain.com/auth/oauth/callback
GITHUB_REDIRECT_URL=https://yourdomain.com/auth/oauth/callback
```

### Update OAuth Provider Settings

1. **Google**: Add production URL to authorized redirect URIs
2. **Facebook**: Add production URL to valid OAuth redirect URIs
3. **GitHub**: Update authorization callback URL

## Troubleshooting

### OAuth Error: Invalid Redirect URI

- Ensure redirect URLs in provider settings match exactly
- Include protocol (http/https)
- No trailing slashes

### OAuth Error: Invalid Client

- Verify client ID and secret are correct
- Check environment variables loaded
- Restart server after changes

### OAuth Succeeds but User Not Created

- Check Supabase RLS policies
- Run `sql/fix_rls_policies.sql`
- Check server logs for errors

## Security Best Practices

1. **Never commit** OAuth secrets to git
2. Use `.env` file (already in `.gitignore`)
3. Rotate secrets regularly
4. Use different credentials for dev/prod
5. Enable 2FA on OAuth provider accounts
6. Monitor OAuth usage in provider dashboards

## Optional: Custom OAuth Buttons

Edit `/templates/auth/login.html` to customize OAuth buttons with your brand colors and icons.

---

**OAuth Setup Complete!** 🎉 Users can now sign in with their preferred provider.



