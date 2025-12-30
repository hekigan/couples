# Database Migrations

This directory contains SQL migration files for incremental database updates.

## How to Apply Migrations in Supabase Dashboard

### Method 1: SQL Editor (Recommended)

1. Open your Supabase project dashboard
2. Navigate to **SQL Editor** in the left sidebar
3. Click **"New query"**
4. Copy and paste the entire contents of the migration file
5. Click **"Run"** or press `Ctrl+Enter` (Windows/Linux) / `Cmd+Enter` (Mac)
6. Check the output panel for success messages

### Method 2: Database Connection

If you prefer using psql or another PostgreSQL client:

```bash
# Connect to your Supabase database
psql "postgresql://postgres:[YOUR-PASSWORD]@[YOUR-PROJECT-REF].supabase.co:5432/postgres"

# Run the migration
\i sql/migrations/001_add_friend_email_invitations.sql
```

## Available Migrations

### 001_add_friend_email_invitations.sql

**Date:** 2025-12-31
**Description:** Adds email-based friend invitation system

**What it does:**
- Creates `friend_email_invitations` table
- Adds indexes for performance
- Creates trigger for `updated_at` column
- Disables RLS (Row Level Security)

**Safe to re-run:** ✅ Yes (uses `IF NOT EXISTS` clauses)

**Prerequisites:**
- Existing `users` table (with `id` column)
- `uuid-ossp` extension enabled
- `update_updated_at_column()` function (optional, skips trigger if missing)

## Migration Status

You can check if a migration has been applied by querying the table:

```sql
-- Check if friend_email_invitations table exists
SELECT EXISTS (
    SELECT FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name = 'friend_email_invitations'
);

-- Count existing email invitations (after migration)
SELECT COUNT(*) FROM friend_email_invitations;
```

## Rollback (if needed)

If you need to rollback the migration:

```sql
-- Drop the table and all related objects
DROP TABLE IF EXISTS friend_email_invitations CASCADE;
```

⚠️ **Warning:** This will permanently delete all email invitation data!

## Best Practices

1. **Backup first:** Always backup your database before running migrations
2. **Test locally:** Test migrations on local Supabase instance first
3. **Read the migration:** Review the SQL before running it
4. **Check output:** Verify the success messages after running
5. **Monitor logs:** Check Supabase logs for any errors

## Getting Supabase Connection String

1. Go to **Settings** → **Database** in Supabase dashboard
2. Find **Connection string** section
3. Copy the **Connection pooling** or **Direct connection** string
4. Replace `[YOUR-PASSWORD]` with your database password

## Troubleshooting

### "relation already exists" error
This means the migration was already applied. Safe to ignore.

### "function does not exist" error for trigger
The migration will skip trigger creation and show a notice. This is safe - the trigger is optional.

### "permission denied" error
Make sure you're using a superuser account or have appropriate permissions.

### Need help?
Check the migration output messages or contact your database administrator.
