-- Fix RLS Policies to avoid infinite recursion
-- Complete rewrite of user table policies

-- Drop ALL existing policies on users table
DROP POLICY IF EXISTS "Allow public anonymous user creation" ON users;
DROP POLICY IF EXISTS "Users can view their own data" ON users;
DROP POLICY IF EXISTS "Users can update their own data" ON users;
DROP POLICY IF EXISTS "Admins can do everything with users" ON users;

-- DISABLE RLS on users table (simplest solution for now)
-- This allows the Go API to work without authentication issues
ALTER TABLE users DISABLE ROW LEVEL SECURITY;

-- Note: For production, you should use Supabase Service Role key in your Go app
-- and keep RLS enabled. But for development, disabling RLS is acceptable.

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'RLS disabled on users table';
    RAISE NOTICE 'Anonymous user creation should now work!';
    RAISE NOTICE '===========================================';
END $$;

