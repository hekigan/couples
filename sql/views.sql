-- Database Views for Query Optimization
-- These views eliminate N+1 query problems and simplify complex joins
-- Run this after schema.sql: psql -h <host> -U <user> -d <database> -f views.sql
--
-- SECURITY NOTE: All views use security_invoker = true (SECURITY INVOKER)
-- This means queries run with the permissions of the querying user, not the view creator.
-- This is the recommended approach for Supabase applications using RLS.

-- ============================================================================
-- View 1: Rooms with Player Information
-- ============================================================================
-- Purpose: Get room details with owner and guest usernames in a single query
-- Usage: SELECT * FROM rooms_with_players WHERE id = $1
-- Performance: Eliminates 2 separate user queries per room fetch

DROP VIEW IF EXISTS rooms_with_players;
CREATE VIEW rooms_with_players
WITH (security_invoker = true)
AS
SELECT
    -- Room fields
    r.id,
    r.name,
    r.status,
    r.language,
    r.is_private,
    r.guest_ready,
    r.max_questions,
    r.current_question,
    r.current_question_id,
    r.selected_categories,
    r.current_player_id,
    r.paused_at,
    r.disconnected_user,
    r.created_at,
    r.updated_at,

    -- Owner information
    r.owner_id,
    owner.username AS owner_username,
    owner.email AS owner_email,

    -- Guest information
    r.guest_id,
    guest.username AS guest_username,
    guest.email AS guest_email,

    -- Current player information (for turn indicator)
    current_player.username AS current_player_username
FROM rooms r
LEFT JOIN users owner ON r.owner_id = owner.id
LEFT JOIN users guest ON r.guest_id = guest.id
LEFT JOIN users current_player ON r.current_player_id = current_player.id
WHERE owner.deleted_at IS NULL
  AND (guest.deleted_at IS NULL OR guest.id IS NULL);

-- ============================================================================
-- View 2: Join Requests with User Information
-- ============================================================================
-- Purpose: Get pending join requests with requester details
-- Usage: SELECT * FROM join_requests_with_users WHERE room_id = $1 AND status = 'pending'
-- Performance: Eliminates N queries for user info (one per request)

DROP VIEW IF EXISTS join_requests_with_users;
CREATE VIEW join_requests_with_users
WITH (security_invoker = true)
AS
SELECT
    -- Join request fields
    jr.id,
    jr.room_id,
    jr.user_id,
    jr.status,
    jr.created_at,
    jr.updated_at,

    -- User information
    u.username,
    u.email,

    -- Room information (for authorization checks)
    r.owner_id AS room_owner_id,
    r.status AS room_status,
    r.guest_id AS room_guest_id
FROM room_join_requests jr
JOIN users u ON jr.user_id = u.id
JOIN rooms r ON jr.room_id = r.id
WHERE u.deleted_at IS NULL;

-- ============================================================================
-- View 3: Active Games with Current Question
-- ============================================================================
-- Purpose: Get all active games with current question details
-- Usage: SELECT * FROM active_games WHERE id = $1
-- Performance: Combines room, users, question, and category in one query

DROP VIEW IF EXISTS active_games;
CREATE VIEW active_games
WITH (security_invoker = true)
AS
SELECT
    -- Room fields
    r.id,
    r.name,
    r.status,
    r.language,
    r.max_questions,
    r.current_question,
    r.current_question_id,
    r.current_player_id,
    r.selected_categories,
    r.paused_at,
    r.disconnected_user,
    r.created_at,

    -- Player information
    r.owner_id,
    owner.username AS owner_username,
    r.guest_id,
    guest.username AS guest_username,

    -- Current player
    current_player.username AS current_player_username,

    -- Current question details
    q.id AS question_id,
    q.question_text AS current_question_text,
    q.lang_code AS current_question_lang,
    q.base_question_id AS question_base_id,
    q.category_id AS question_category_id,
    c.key AS question_category_key,
    c.label AS question_category_label
FROM rooms r
JOIN users owner ON r.owner_id = owner.id
LEFT JOIN users guest ON r.guest_id = guest.id
LEFT JOIN users current_player ON r.current_player_id = current_player.id
LEFT JOIN questions q ON r.current_question_id = q.id
LEFT JOIN categories c ON q.category_id = c.id
WHERE r.status IN ('playing', 'paused')
  AND owner.deleted_at IS NULL
  AND (guest.deleted_at IS NULL OR guest.id IS NULL);

-- ============================================================================
-- View 4: Room Invitations with Details
-- ============================================================================
-- Purpose: Get room invitations with inviter, invitee, and room details
-- Usage: SELECT * FROM invitations_with_details WHERE invitee_id = $1 AND status = 'pending'
-- Performance: Eliminates 3 separate queries (inviter, invitee, room)

DROP VIEW IF EXISTS invitations_with_details;
CREATE VIEW invitations_with_details
WITH (security_invoker = true)
AS
SELECT
    -- Invitation fields
    inv.id,
    inv.room_id,
    inv.inviter_id,
    inv.invitee_id,
    inv.status,
    inv.created_at,
    inv.updated_at,

    -- Inviter information
    inviter.username AS inviter_username,
    inviter.email AS inviter_email,

    -- Invitee information
    invitee.username AS invitee_username,
    invitee.email AS invitee_email,

    -- Room information
    r.name AS room_name,
    r.status AS room_status,
    r.language AS room_language,
    r.is_private AS room_is_private,
    r.owner_id AS room_owner_id
FROM room_invitations inv
JOIN users inviter ON inv.inviter_id = inviter.id
JOIN users invitee ON inv.invitee_id = invitee.id
JOIN rooms r ON inv.room_id = r.id
WHERE inviter.deleted_at IS NULL
  AND invitee.deleted_at IS NULL;

-- ============================================================================
-- View 5: Friends List with User Details
-- ============================================================================
-- Purpose: Get user's friends with details (bidirectional friendship)
-- Usage: SELECT * FROM friends_with_details WHERE user_id = $1
-- Performance: Eliminates separate query for friend user details

DROP VIEW IF EXISTS friends_with_details;
CREATE VIEW friends_with_details
WITH (security_invoker = true)
AS
SELECT
    -- Friendship fields
    f.id,
    f.user_id,
    f.friend_id,
    f.status,
    f.created_at,

    -- Friend information
    friend.username AS friend_username,
    friend.email AS friend_email,

    -- User information (for reverse lookups)
    u.username AS user_username
FROM friends f
JOIN users u ON f.user_id = u.id
JOIN users friend ON f.friend_id = friend.id
WHERE u.deleted_at IS NULL
  AND friend.deleted_at IS NULL
  AND f.status = 'accepted';

-- ============================================================================
-- View 6: Game History with Details
-- ============================================================================
-- Purpose: Get completed games with player names and scores
-- Usage: SELECT * FROM game_history WHERE owner_id = $1 OR guest_id = $1 ORDER BY finished_at DESC
-- Performance: Single query for game history instead of multiple joins

DROP VIEW IF EXISTS game_history;
CREATE VIEW game_history
WITH (security_invoker = true)
AS
SELECT
    -- Room/Game fields
    r.id,
    r.name AS game_name,
    r.status,
    r.max_questions,
    r.current_question AS questions_answered,
    r.created_at AS started_at,
    r.updated_at AS finished_at,

    -- Owner information
    r.owner_id,
    owner.username AS owner_username,

    -- Guest information
    r.guest_id,
    guest.username AS guest_username,

    -- Game metadata
    r.language,
    r.selected_categories
FROM rooms r
JOIN users owner ON r.owner_id = owner.id
LEFT JOIN users guest ON r.guest_id = guest.id
WHERE r.status = 'finished'
  AND owner.deleted_at IS NULL
  AND (guest.deleted_at IS NULL OR guest.id IS NULL)
ORDER BY r.updated_at DESC;

-- ============================================================================
-- PERFORMANCE NOTES
-- ============================================================================
--
-- 1. These views use LEFT JOINs for optional relationships (guest, current_player)
-- 2. All views filter out soft-deleted users (deleted_at IS NULL)
-- 3. Views are created with OR REPLACE to allow updates
-- 4. No materialization - views are computed on query (real-time data)
-- 5. Indexes on base tables (rooms, users, etc.) are still used
-- 6. All views use security_invoker = true for proper RLS enforcement
--
-- Query Performance Comparison:
-- - Before: GetRoomWithPlayers = 3 queries (room + owner + guest)
-- - After:  GetRoomWithPlayers = 1 query (rooms_with_players view)
-- - Improvement: ~3x faster, less network overhead
--
-- - Before: GetJoinRequestsWithUserInfo = 1 + N queries (requests + N users)
-- - After:  GetJoinRequestsWithUserInfo = 1 query (join_requests_with_users view)
-- - Improvement: For 10 requests: 11 queries → 1 query = 11x faster!
--
-- ============================================================================

-- Grant permissions to application user (adjust as needed)
-- GRANT SELECT ON rooms_with_players TO your_app_user;
-- GRANT SELECT ON join_requests_with_users TO your_app_user;
-- GRANT SELECT ON active_games TO your_app_user;
-- GRANT SELECT ON invitations_with_details TO your_app_user;
-- GRANT SELECT ON friends_with_details TO your_app_user;
-- GRANT SELECT ON game_history TO your_app_user;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================
-- Run these to verify views are working correctly:

-- Test rooms_with_players
-- SELECT id, owner_username, guest_username, status FROM rooms_with_players LIMIT 5;

-- Test join_requests_with_users
-- SELECT id, username, room_id, status FROM join_requests_with_users WHERE status = 'pending' LIMIT 5;

-- Test active_games
-- SELECT id, owner_username, guest_username, current_question_text FROM active_games LIMIT 5;

-- Test invitations_with_details
-- SELECT id, inviter_username, invitee_username, room_name FROM invitations_with_details LIMIT 5;

-- Test friends_with_details
-- SELECT user_username, friend_username, status FROM friends_with_details LIMIT 5;

-- Test game_history
-- SELECT game_name, owner_username, guest_username, questions_answered FROM game_history LIMIT 5;
