# Email Troubleshooting Guide

## Quick Test: Verify Email Setup

I've added a test endpoint that will immediately confirm if your email service is working.

### Step 1: Start Your Server

```bash
make dev
# or
./server
```

### Step 2: Test Email Sending

**Open your browser or use curl:**

```bash
# Test with default test email (delivered@resend.dev)
curl http://localhost:8080/api/test-email

# Test with your own email
curl http://localhost:8080/api/test-email?to=your@email.com
```

**Or visit in browser:**
- http://localhost:8080/api/test-email
- http://localhost:8080/api/test-email?to=your@email.com

### Step 3: Check the Response

#### ✅ Success Response

```json
{
  "success": true,
  "message": "Test email sent successfully! Check your inbox (or Resend dashboard).",
  "details": {
    "email_id": "abc123-def456-ghi789",
    "from": "noreply@yourdomain.com",
    "to": "delivered@resend.dev",
    "subject": "Test Email from Couple Card Game - 14:35:22",
    "is_development": true,
    "timestamp": "2025-12-31 14:35:22"
  },
  "next_steps": [
    "Check Resend dashboard: https://resend.com/emails",
    "Look for email ID: abc123-def456-ghi789",
    "Recipient: delivered@resend.dev",
    "Email should arrive within seconds"
  ]
}
```

**What this means:**
- ✅ Resend API key is valid
- ✅ Email service is configured correctly
- ✅ Email was sent successfully
- 📧 Check https://resend.com/emails to see the email

#### ❌ Error Response

```json
{
  "success": false,
  "error": "Post \"https://api.resend.com/emails\": unauthorized",
  "message": "Failed to send test email. Check your RESEND_API_KEY and configuration.",
  "config": {
    "from": "noreply@yourdomain.com",
    "to": "delivered@resend.dev",
    "api_key_set": true,
    "is_development": true
  }
}
```

**What this means:**
- ❌ There's a problem with your configuration
- Check the `error` field for details

### Step 4: View Email in Resend Dashboard

1. Go to https://resend.com/emails
2. Look for the email with the `email_id` from the response
3. Click on it to see:
   - Full HTML preview
   - Delivery status
   - Timestamp
   - Tags

## Common Issues & Solutions

### Issue 1: "Email service not configured"

**Error:**
```json
{
  "error": "Email service not configured",
  "message": "Please set RESEND_API_KEY, EMAIL_FROM, and APP_BASE_URL in .env file"
}
```

**Solution:**
1. Check your `.env` file exists
2. Verify these variables are set:
   ```env
   RESEND_API_KEY=re_your_api_key_here
   EMAIL_FROM=noreply@yourdomain.com
   APP_BASE_URL=http://localhost:8080
   ```
3. Restart your server: `make dev`

### Issue 2: "unauthorized" or "invalid API key"

**Error:**
```json
{
  "error": "Post \"https://api.resend.com/emails\": unauthorized"
}
```

**Solution:**
1. Check your Resend API key is correct
2. Log in to https://resend.com/api-keys
3. Verify the key starts with `re_`
4. Copy the key exactly (no extra spaces)
5. Update `.env` file
6. Restart server

### Issue 3: "domain not verified"

**Error:**
```json
{
  "error": "The gmail.com domain is not verified"
}
```

**Solution (Development - Testing without a domain):**
1. Use Resend's pre-verified test sender: `onboarding@resend.dev`
2. Update your `.env` file:
   ```env
   EMAIL_FROM=onboarding@resend.dev
   ```
3. Restart your server: `make dev`
4. No domain verification needed for testing!

**Important:** You cannot use personal email addresses like `yourname@gmail.com` as the sender. Resend requires either:
- A verified domain you own
- Their test sender address: `onboarding@resend.dev`

**Solution (Production - Using your own domain):**
1. Go to https://resend.com/domains
2. Add your domain (e.g., `yourdomain.com`)
3. Configure DNS records (SPF, DKIM, DMARC)
4. Wait for verification (usually minutes)
5. Update `EMAIL_FROM` in `.env` to use verified domain (e.g., `noreply@yourdomain.com`)

### Issue 4: "context canceled"

**Error in server logs:**
```
❌ Failed to send email: context canceled
```

**Solution:**
✅ **Already fixed!** We're now using `context.Background()` for async email sending.

### Issue 5: No emails appear in Resend dashboard

**Possible causes:**

1. **Wrong Resend account**
   - Make sure you're logged into the correct Resend account
   - Check the API key belongs to this account

2. **API rate limit exceeded**
   - Free tier: 100 emails/day, 3,000/month
   - Check https://resend.com/overview for usage

3. **Email stuck in queue**
   - Usually arrives within seconds
   - Check Resend status: https://resend.com/status

4. **API key revoked**
   - Regenerate API key at https://resend.com/api-keys

## Debugging Checklist

Go through this checklist:

- [ ] `.env` file exists in project root
- [ ] `RESEND_API_KEY` is set (starts with `re_`)
- [ ] `EMAIL_FROM` is set
- [ ] `APP_BASE_URL` is set
- [ ] Server is running (`make dev`)
- [ ] Test endpoint returns success: http://localhost:8080/api/test-email
- [ ] Email appears in Resend dashboard: https://resend.com/emails
- [ ] No errors in server console logs

## Environment Variables Template

Copy this to your `.env` file:

```env
# Environment
ENV=development

# Resend Email Service
RESEND_API_KEY=re_YOUR_API_KEY_HERE
EMAIL_FROM=onboarding@resend.dev  # Use this for testing (pre-verified by Resend)
APP_BASE_URL=http://localhost:8080

# Supabase (your existing config)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Session
SESSION_SECRET=your-session-secret-min-32-chars
```

## Getting a Resend API Key

1. Sign up at https://resend.com
2. Click "API Keys" in the left sidebar
3. Click "Create API Key"
4. Give it a name (e.g., "Development")
5. Copy the key (starts with `re_`)
6. Paste into `.env` file

**Free tier includes:**
- 3,000 emails/month
- 100 emails/day
- No credit card required
- Test emails (delivered@resend.dev)

## Still Not Working?

### Enable Debug Logging

Check your server console for these messages:

**On server start:**
```
📧 Email service running in DEVELOPMENT mode - using test emails (delivered@resend.dev)
✅ Email service initialized
```

**When sending emails:**
```
📧 [DEV] Sending join invitation email to TEST address (original: test@example.com, sender: Laurent, token: UiDf3b0p6emSgt-q...)
```

### Manual cURL Test

Test Resend API directly:

```bash
curl -X POST 'https://api.resend.com/emails' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{
    "from": "noreply@yourdomain.com",
    "to": ["delivered@resend.dev"],
    "subject": "Test",
    "html": "<p>Test email</p>"
  }'
```

**Success response:**
```json
{
  "id": "abc123-def456-ghi789"
}
```

**Error response:**
```json
{
  "message": "Invalid API key",
  "name": "validation_error"
}
```

## Next Steps After Testing

Once the test email works:

1. ✅ Test email endpoint confirms everything works
2. 🧪 Try sending friend invitation via the UI
3. 📧 Check Resend dashboard for the invitation email
4. 🎉 Your email system is fully operational!

## Support Resources

- Resend Documentation: https://resend.com/docs
- Resend Status Page: https://resend.com/status
- Resend API Keys: https://resend.com/api-keys
- Resend Domains: https://resend.com/domains
- Email Logs: https://resend.com/emails
