-- Fix RLS policies to allow anonymous user creation
-- This script updates the users table policies to support guest/anonymous users

-- First, drop existing policies
DROP POLICY IF EXISTS "Allow public anonymous user creation" ON users;
DROP POLICY IF EXISTS "Users can view their own data" ON users;
DROP POLICY IF EXISTS "Users can update their own data" ON users;
DROP POLICY IF EXISTS "Admins can do everything with users" ON users;

-- Create new policies that support anonymous users
CREATE POLICY "Allow public anonymous user creation"
ON users FOR INSERT
TO public
WITH CHECK (is_anonymous = TRUE);

CREATE POLICY "Users can view their own data"
ON users FOR SELECT
TO public
USING (
  auth.uid() = id OR is_anonymous = TRUE
);

CREATE POLICY "Users can update their own data"
ON users FOR UPDATE
TO public
USING (auth.uid() = id)
WITH CHECK (auth.uid() = id);

CREATE POLICY "Admins can do everything with users"
ON users FOR ALL
TO public
USING (
  EXISTS (
    SELECT 1 FROM auth.users
    WHERE auth.users.id = auth.uid()
    AND auth.users.raw_user_meta_data->>'role' = 'admin'
  )
);

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'RLS policies updated for anonymous users';
    RAISE NOTICE 'Anonymous user creation should now work!';
    RAISE NOTICE '===========================================';
END $$;

