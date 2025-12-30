# Email Testing Guide

This guide explains how email testing works in the Couple Card Game application using Resend's test email feature.

## Overview

The application automatically detects the environment and uses **test emails** in development mode to avoid sending emails to real recipients during testing.

## How It Works

### Development Mode (ENV=development)

When `ENV=development` is set in your `.env` file, all emails are:

1. **Redirected to test address**: `delivered@resend.dev` (Resend's test email)
2. **Tagged with metadata**: Original recipient, email type, environment
3. **Logged to console**: Shows original recipient and email details

### Production Mode (ENV=production or unset)

Emails are sent to **real recipients** as specified in the code.

## Email Tags (Development Mode)

All development emails include tags to help you track and filter them in the Resend dashboard:

| Tag Name | Description | Example Values |
|----------|-------------|----------------|
| `environment` | Deployment environment | `development` |
| `email_type` | Type of email sent | `friend_request_existing`, `friend_invite_new_user` |
| `original_recipient` | Who would have received this email in production | `user@example.com` |
| `signup_token` | Token prefix for join invitations (first 16 chars) | `abc123def456...` |

## Console Logs (Development Mode)

When emails are sent in development, you'll see logs like:

```
📧 Email service running in DEVELOPMENT mode - using test emails (delivered@resend.dev)
📧 [DEV] Sending friend request email to TEST address (original: alice@example.com, sender: Bob)
📧 [DEV] Sending join invitation email to TEST address (original: charlie@example.com, sender: Bob, token: Xy7aB9cD3eFgH1...)
```

## Setup

### 1. Environment Variables

Add to your `.env` file:

```env
# Environment mode
ENV=development

# Resend configuration
RESEND_API_KEY=re_your_api_key_here
EMAIL_FROM=noreply@yourdomain.com
APP_BASE_URL=http://localhost:8080
```

### 2. Get Resend API Key

1. Sign up at https://resend.com
2. Navigate to **API Keys** in the dashboard
3. Create a new API key
4. Copy and paste into `.env`

### 3. Verify Domain (Optional for Development)

For development, you can use Resend's test emails without verifying a domain. However, for production:

1. Go to **Domains** in Resend dashboard
2. Add your domain
3. Configure DNS records as shown
4. Wait for verification

## Testing Friend Invitations

### Test Scenario 1: Invite Existing User

1. Start your dev server: `make dev`
2. Navigate to http://localhost:8080/friends/add
3. Select "Email Address" from dropdown
4. Enter any email: `test@example.com`
5. Click "Send Invitation"

**What happens:**
- Email sent to `delivered@resend.dev` (not `test@example.com`)
- Console shows: `📧 [DEV] Sending friend request email to TEST address (original: test@example.com, sender: YourName)`
- In Resend dashboard, you can filter by tag `email_type=friend_request_existing`

### Test Scenario 2: Invite New User

1. Navigate to http://localhost:8080/friends/add
2. Select "Email Address" from dropdown
3. Enter email for non-existent user: `newuser@example.com`
4. Click "Send Invitation"

**What happens:**
- Email sent to `delivered@resend.dev` (not `newuser@example.com`)
- Console shows: `📧 [DEV] Sending join invitation email to TEST address (original: newuser@example.com, sender: YourName, token: ...)`
- Email includes signup link with auto-accept token
- In Resend dashboard, filter by tag `email_type=friend_invite_new_user`

## Viewing Test Emails in Resend Dashboard

### Method 1: Email List

1. Go to https://resend.com/emails
2. All development emails will appear with recipient `delivered@resend.dev`
3. Click on any email to see:
   - Full HTML preview
   - Tags (environment, email_type, original_recipient)
   - Delivery status
   - Timestamp

### Method 2: Filter by Tags

To find specific emails:

```
# Find all development emails
Tag: environment = development

# Find friend request emails only
Tag: email_type = friend_request_existing

# Find emails for specific recipient
Tag: original_recipient = alice@example.com
```

### Method 3: Search

Use the search bar in Resend dashboard:
- Search by email subject
- Search by tag values
- Search by date range

## Resend Test Email Addresses

Resend provides several test email addresses for different scenarios:

| Email Address | Behavior | Use Case |
|---------------|----------|----------|
| `delivered@resend.dev` | Successfully delivered (200 OK) | **Default for our app** - Normal email flow |
| `bounced@resend.dev` | Hard bounce (rejected) | Test error handling for invalid addresses |
| `complained@resend.dev` | Spam complaint | Test spam handling |
| `unsubscribed@resend.dev` | Unsubscribed user | Test unsubscribe flow (if implemented) |

We use `delivered@resend.dev` by default to simulate successful delivery.

## Switching to Production

When deploying to production:

1. Update `.env`:
   ```env
   ENV=production
   ```

2. Verify all emails go to **real recipients**

3. Monitor Resend dashboard for delivery rates

4. Check for bounces and complaints

## Email Content Preview

### Friend Request Email (Existing User)

**Subject:** `[SenderName] sent you a friend request`

**Content:**
- Friendly greeting
- Sender's name prominently displayed
- Call-to-action button: "View Friend Requests"
- Link to `/friends` page

### Join Invitation Email (New User)

**Subject:** `[SenderName] invited you to join Couple Card Game`

**Content:**
- Invitation from sender
- App description and benefits
- List of features (questions, categories, multiplayer, i18n)
- Call-to-action button: "Accept Invitation & Sign Up"
- Signup link with auto-accept token
- Expiration notice (7 days)

## Troubleshooting

### Emails not appearing in Resend dashboard

**Problem:** Sent emails don't show up in dashboard

**Solutions:**
1. Check Resend API key is valid
2. Verify `ENV=development` is set
3. Check console logs for error messages
4. Confirm Resend API limits (3,000/month on free tier)

### "Failed to send email" errors

**Problem:** Console shows `❌ Failed to send email`

**Solutions:**
1. Verify `RESEND_API_KEY` is correct in `.env`
2. Check Resend API status: https://resend.com/status
3. Review API limits (100 emails/day on free tier)
4. Check domain verification status

### Tags not showing in dashboard

**Problem:** Tags are missing in email details

**Solutions:**
1. Ensure `ENV=development` is set (tags only added in dev mode)
2. Check Go version compatibility with Resend SDK
3. Verify Resend SDK version: `go list -m github.com/resend/resend-go/v2`

### Original recipient not visible

**Problem:** Can't see who the email was intended for

**Solutions:**
1. Check console logs - shows original recipient
2. View email tags in Resend dashboard
3. Look for `original_recipient` tag

## Best Practices

### Development
- ✅ Always use `ENV=development` locally
- ✅ Review emails in Resend dashboard
- ✅ Test both invitation types (existing user, new user)
- ✅ Monitor console logs for debugging

### Production
- ✅ Set `ENV=production` in production environment
- ✅ Verify domain before going live
- ✅ Monitor bounce rates
- ✅ Set up error alerts for failed emails

### Testing
- ✅ Test with various email formats
- ✅ Verify links work (signup token, friend requests)
- ✅ Check email rendering on different clients (Gmail, Outlook, etc.)
- ✅ Test error scenarios (invalid API key, rate limits)

## Rate Limits

**Resend Free Tier:**
- 3,000 emails/month
- 100 emails/day
- No credit card required

**Recommendations:**
- Use test emails in development (don't waste quota)
- Monitor usage in Resend dashboard
- Upgrade to paid plan for production (higher limits, SLA)

## Additional Resources

- [Resend Documentation](https://resend.com/docs)
- [Send Test Emails Guide](https://resend.com/docs/dashboard/emails/send-test-emails)
- [Resend Go SDK](https://github.com/resend/resend-go)
- [Email Tags Documentation](https://resend.com/docs/dashboard/emails/tags)

## Support

If you encounter issues:

1. Check Resend status page: https://resend.com/status
2. Review Resend documentation: https://resend.com/docs
3. Check application logs in console
4. Verify environment variables are set correctly
