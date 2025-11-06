-- ============================================================================
-- DROP ALL TABLES - Complete Database Reset
-- ============================================================================
-- ⚠️  WARNING: This script will DELETE ALL DATA
-- Only run this if you want to completely reset the database
-- 
-- After running this, you can recreate everything with:
--   psql -f schema.sql
--   psql -f seed.sql (optional)
-- ============================================================================

-- Drop all tables (order matters due to foreign keys)
DROP TABLE IF EXISTS answers CASCADE;
DROP TABLE IF EXISTS question_history CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS room_invitations CASCADE;
DROP TABLE IF EXISTS room_join_requests CASCADE;
DROP TABLE IF EXISTS rooms CASCADE;
DROP TABLE IF EXISTS questions CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS friend_requests CASCADE;
DROP TABLE IF EXISTS friends CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS translations CASCADE;

-- Drop extensions (optional)
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
DROP EXTENSION IF EXISTS "pg_trgm" CASCADE;

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'All tables dropped successfully';
    RAISE NOTICE 'Database is now empty';
    RAISE NOTICE 'Run schema.sql to recreate tables';
    RAISE NOTICE '===========================================';
END $$;

