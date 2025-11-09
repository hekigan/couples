-- Database Indexes for Performance Optimization
-- These indexes improve query performance for common access patterns
-- Run this after schema.sql and views.sql

-- ============================================================================
-- ROOMS TABLE INDEXES
-- ============================================================================

-- Index for finding rooms by status (for listing active/available rooms)
CREATE INDEX IF NOT EXISTS idx_rooms_status
ON rooms(status)
WHERE status IN ('waiting', 'ready', 'playing');

-- Index for finding rooms by owner (for "my rooms" queries)
CREATE INDEX IF NOT EXISTS idx_rooms_owner_id
ON rooms(owner_id)
WHERE owner_id IS NOT NULL;

-- Index for finding rooms by guest (for "rooms I'm in" queries)
CREATE INDEX IF NOT EXISTS idx_rooms_guest_id
ON rooms(guest_id)
WHERE guest_id IS NOT NULL;

-- Composite index for finding available rooms (public, waiting status)
CREATE INDEX IF NOT EXISTS idx_rooms_public_waiting
ON rooms(is_private, status, created_at DESC)
WHERE is_private = FALSE AND status = 'waiting';

-- Index for finding active games by current player (for turn notifications)
CREATE INDEX IF NOT EXISTS idx_rooms_current_player
ON rooms(current_player_id, status)
WHERE current_player_id IS NOT NULL AND status = 'playing';

-- ============================================================================
-- ROOM_JOIN_REQUESTS TABLE INDEXES
-- ============================================================================

-- Index for finding pending requests by room (most common query)
CREATE INDEX IF NOT EXISTS idx_join_requests_room_status
ON room_join_requests(room_id, status, created_at DESC)
WHERE status = 'pending';

-- Index for finding user's pending requests (to prevent duplicates)
CREATE INDEX IF NOT EXISTS idx_join_requests_user_room
ON room_join_requests(user_id, room_id, status)
WHERE status = 'pending';

-- Index for finding all requests by user (user's request history)
CREATE INDEX IF NOT EXISTS idx_join_requests_user_id
ON room_join_requests(user_id, created_at DESC);

-- ============================================================================
-- ROOM_INVITATIONS TABLE INDEXES
-- ============================================================================

-- Index for finding pending invitations for a user
CREATE INDEX IF NOT EXISTS idx_invitations_invitee_status
ON room_invitations(invitee_id, status, created_at DESC)
WHERE status = 'pending';

-- Index for finding invitations by room
CREATE INDEX IF NOT EXISTS idx_invitations_room_id
ON room_invitations(room_id, status)
WHERE status = 'pending';

-- Index for finding invitations sent by a user
CREATE INDEX IF NOT EXISTS idx_invitations_inviter_id
ON room_invitations(inviter_id, created_at DESC);

-- ============================================================================
-- ANSWERS TABLE INDEXES
-- ============================================================================

-- Index for finding answers by room (for game history/review)
CREATE INDEX IF NOT EXISTS idx_answers_room_id
ON answers(room_id, created_at);

-- Index for finding answers by user (user's answer history)
CREATE INDEX IF NOT EXISTS idx_answers_user_id
ON answers(user_id, created_at DESC);

-- Index for finding answers by question (for analysis)
CREATE INDEX IF NOT EXISTS idx_answers_question_id
ON answers(question_id);

-- Composite index for finding room answers by user (most common query)
CREATE INDEX IF NOT EXISTS idx_answers_room_user
ON answers(room_id, user_id, created_at);

-- ============================================================================
-- QUESTIONS TABLE INDEXES
-- ============================================================================

-- Index for finding questions by category (for filtered question selection)
CREATE INDEX IF NOT EXISTS idx_questions_category_lang
ON questions(category_id, lang_code)
WHERE category_id IS NOT NULL;

-- Index for finding questions by language
CREATE INDEX IF NOT EXISTS idx_questions_lang_code
ON questions(lang_code);

-- ============================================================================
-- FRIENDS TABLE INDEXES
-- ============================================================================

-- Index for finding user's friends (bidirectional)
CREATE INDEX IF NOT EXISTS idx_friends_user_status
ON friends(user_id, status)
WHERE status = 'accepted';

-- Index for finding friend requests received
CREATE INDEX IF NOT EXISTS idx_friends_friend_status
ON friends(friend_id, status)
WHERE status = 'pending';

-- Composite index for checking friendship existence
CREATE INDEX IF NOT EXISTS idx_friends_both_users
ON friends(user_id, friend_id, status);

-- ============================================================================
-- USERS TABLE INDEXES
-- ============================================================================

-- Index for finding users by username (login, search)
CREATE INDEX IF NOT EXISTS idx_users_username
ON users(username)
WHERE deleted_at IS NULL;

-- Index for finding users by email (login)
CREATE INDEX IF NOT EXISTS idx_users_email
ON users(email)
WHERE deleted_at IS NULL;

-- Index for finding active users (exclude soft-deleted)
CREATE INDEX IF NOT EXISTS idx_users_deleted_at
ON users(deleted_at)
WHERE deleted_at IS NULL;

-- ============================================================================
-- NOTIFICATIONS TABLE INDEXES
-- ============================================================================

-- Index for finding unread notifications for a user
-- Note: Column is called 'read' not 'is_read' in schema.sql
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
ON notifications(user_id, read, created_at DESC)
WHERE read = FALSE;

-- Index for finding all user notifications (notification center)
CREATE INDEX IF NOT EXISTS idx_notifications_user_created
ON notifications(user_id, created_at DESC);

-- ============================================================================
-- PERFORMANCE ANALYSIS
-- ============================================================================

-- To analyze query performance with these indexes, use EXPLAIN ANALYZE:
--
-- Example 1: Find pending join requests for a room
-- EXPLAIN ANALYZE
-- SELECT * FROM join_requests_with_users
-- WHERE room_id = '123e4567-e89b-12d3-a456-426614174000'
-- AND status = 'pending';
--
-- Expected: Index Scan using idx_join_requests_room_status
--
-- Example 2: Find user's active games
-- EXPLAIN ANALYZE
-- SELECT * FROM active_games
-- WHERE owner_id = '123e4567-e89b-12d3-a456-426614174000'
-- OR guest_id = '123e4567-e89b-12d3-a456-426614174000';
--
-- Expected: Bitmap Index Scan using idx_rooms_owner_id and idx_rooms_guest_id
--
-- Example 3: Find user's friends
-- EXPLAIN ANALYZE
-- SELECT * FROM friends_with_details
-- WHERE user_id = '123e4567-e89b-12d3-a456-426614174000';
--
-- Expected: Index Scan using idx_friends_user_status

-- ============================================================================
-- INDEX MAINTENANCE
-- ============================================================================

-- Rebuild indexes if needed (after bulk operations)
-- REINDEX TABLE rooms;
-- REINDEX TABLE room_join_requests;
-- REINDEX TABLE answers;

-- Analyze tables to update statistics (improves query planner decisions)
-- ANALYZE rooms;
-- ANALYZE room_join_requests;
-- ANALYZE room_invitations;
-- ANALYZE answers;
-- ANALYZE users;
-- ANALYZE friends;

-- ============================================================================
-- MONITORING
-- ============================================================================

-- Check index usage (identify unused indexes)
-- SELECT
--     schemaname,
--     tablename,
--     indexname,
--     idx_scan as index_scans,
--     idx_tup_read as tuples_read,
--     idx_tup_fetch as tuples_fetched
-- FROM pg_stat_user_indexes
-- WHERE schemaname = 'public'
-- ORDER BY idx_scan ASC;

-- Check table sizes (identify large tables needing optimization)
-- SELECT
--     tablename,
--     pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
-- FROM pg_tables
-- WHERE schemaname = 'public'
-- ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- ============================================================================
-- NOTES
-- ============================================================================
--
-- 1. Partial indexes (WHERE clauses) are used for frequently queried subsets
--    - Smaller index size
--    - Faster index scans
--    - Only applicable queries use them
--
-- 2. Composite indexes follow the "leftmost prefix" rule
--    - Can be used for queries on (col1), (col1, col2), or (col1, col2, col3)
--    - Cannot be used for queries on just (col2) or (col3)
--
-- 3. Index on (created_at DESC) supports ORDER BY created_at DESC queries
--
-- 4. Indexes have overhead:
--    - Storage space
--    - INSERT/UPDATE/DELETE slower (index must be updated)
--    - Choose indexes based on read vs write patterns
--
-- 5. PostgreSQL automatically creates indexes for:
--    - PRIMARY KEY constraints (already in schema.sql)
--    - UNIQUE constraints (already in schema.sql)
--
-- 6. These indexes complement the existing schema indexes
--
-- ============================================================================
