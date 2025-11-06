# SQL Files

This folder contains all SQL scripts for the Couple Card Game database.

## 📁 Files Overview

### Core Database Files

#### `schema.sql` (11 KB)
**Purpose:** Complete database schema definition  
**When to run:** Initial setup or fresh database creation  
**Contains:**
- Table definitions (users, categories, questions, rooms, etc.)
- Indexes for performance
- Row Level Security (RLS) policies
- Triggers for automatic timestamps

**Usage:**
```sql
-- Run in Supabase SQL Editor
-- Creates all tables and relationships
```

#### `seed.sql` (7.6 KB)
**Purpose:** Sample data for development and testing  
**When to run:** After running schema.sql  
**Contains:**
- Sample categories (Relationship, Fun, Deep, etc.)
- ~30 sample questions in 3 languages (EN, FR, ES)

**Usage:**
```sql
-- Run in Supabase SQL Editor after schema.sql
-- Provides initial content for testing
```

---

### Migration Files

#### `migration_v2.sql` (3 KB)
**Purpose:** Updates existing database with new features  
**When to run:** When upgrading from v1 to v2  
**Contains:**
- Creates `room_join_requests` table
- Adds indexes and RLS policies
- Safe to run multiple times (uses IF NOT EXISTS)

**Usage:**
```sql
-- Run if you already have a database and need to add:
-- - Room join request functionality
```

---

### Fix Scripts

#### `fix_rls_policies.sql` (1 KB)
**Purpose:** Fixes Row Level Security infinite recursion  
**When to run:** If getting "infinite recursion detected" errors  
**Contains:**
- Removes problematic RLS policies
- Disables RLS on users table for development

**Usage:**
```sql
-- Run if anonymous user creation fails
-- Allows API access without authentication issues
```

#### `fix_anonymous_users.sql` (1.3 KB)
**Purpose:** Allows public creation of anonymous users  
**When to run:** If "Play as Guest" functionality doesn't work  
**Contains:**
- Updated RLS policies for anonymous user support
- Allows unauthenticated user creation

**Note:** If this doesn't work, use `fix_rls_policies.sql` instead

---

### Utility Scripts

#### `drop_all.sql` (746 B)
**Purpose:** Complete database reset  
**When to run:** ⚠️ **WARNING** - Only for complete fresh start  
**Contains:**
- DROP TABLE commands for all tables
- Removes all data

**⚠️ DANGER:**
```sql
-- This will DELETE ALL DATA
-- Only run if you want to start completely fresh
-- Then run schema.sql and seed.sql again
```

---

## 🚀 Setup Order

### For New Database:
1. `schema.sql` - Create all tables
2. `seed.sql` - Add sample data
3. Done! ✅

### For Existing Database (Upgrade):
1. `migration_v2.sql` - Add new features
2. `fix_rls_policies.sql` - Fix authentication issues (if needed)

### For Complete Reset:
1. `drop_all.sql` - ⚠️ Delete everything
2. `schema.sql` - Recreate tables
3. `seed.sql` - Add sample data

---

## 📝 Notes

- All scripts are **idempotent** where possible (safe to run multiple times)
- Run in **Supabase SQL Editor**: https://app.supabase.com
- Test in **development** before running in production
- Always **backup** before running drop_all.sql

---

## 🔗 Related Documentation

- [SETUP.md](../docs/SETUP.md) - Complete setup guide
- [PROJECT_STATUS.md](../docs/PROJECT_STATUS.md) - Implementation status
- [START_HERE.md](../START_HERE.md) - Quick start guide

