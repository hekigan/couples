# Testing Guide for Database View Optimizations

## Overview
This guide will help you test the database view optimizations that were implemented in Phase 1 of the HTMX + Supabase refactoring.

## Pre-Testing: Clean Up Problematic Data

Before testing, run this SQL in Supabase to clean up any duplicate or problematic rooms:

```sql
-- Check for duplicate room IDs (shouldn't happen, but just in case)
SELECT id, COUNT(*) as count
FROM rooms
GROUP BY id
HAVING COUNT(*) > 1;

-- If you find any duplicates, keep only the most recent one:
-- DELETE FROM rooms
-- WHERE id IN (
--     SELECT id FROM (
--         SELECT id, ROW_NUMBER() OVER (PARTITION BY id ORDER BY created_at DESC) as rn
--         FROM rooms
--     ) t WHERE rn > 1
-- );

-- Clean up rooms with invalid owner references
DELETE FROM rooms
WHERE owner_id NOT IN (SELECT id FROM users WHERE deleted_at IS NULL);

-- Clean up rooms with invalid guest references (where guest was deleted)
UPDATE rooms
SET guest_id = NULL
WHERE guest_id IS NOT NULL
  AND guest_id NOT IN (SELECT id FROM users WHERE deleted_at IS NULL);
```

## Test 1: RoomHandler Optimization (Room Lobby Page)

**What changed:** `RoomHandler` now uses `GetRoomWithPlayers()` which fetches room + owner + guest in **1 query** instead of **3 queries**.

### Test Steps:
1. Start the server: `make run`
2. Log in to the application
3. Create a new room
4. Navigate to the room lobby page

### What to Look For:
- **Server logs** should show:
  ```
  📊 Room with players found: <room-id> (owner: <username>, guest: <nil>, status: waiting) - single query via view
  ```
- **Page should load** without errors
- **Owner username** should display correctly
- **Guest username** should display correctly (if guest joins)

### Performance Check:
- Before: Server would log 3 separate database queries
- After: Server logs only 1 query via the view

---

## Test 2: ListRoomsHandler Optimization (Rooms List Page)

**What changed:** `ListRoomsHandler` now uses `GetRoomsByUserIDWithPlayers()` which eliminates N+1 queries.

### Test Steps:
1. Create 3-5 rooms (as owner)
2. Join 2-3 other rooms (as guest)
3. Navigate to `/game/rooms`

### What to Look For:
- **Server logs** should show:
  ```
  📊 Found X rooms with player info for user <user-id> (owner: Y, guest: Z) - using view
  ```
- **Page should display** all your rooms correctly
- **Other player usernames** should appear for each room

### Performance Check:
- **Before:** For 10 rooms with players: 12 queries (2 for rooms + 10 for user lookups)
- **After:** 2 queries only (owner rooms + guest rooms, both via view)

### How to Verify Performance:
Look for the absence of multiple individual user queries in the logs. You should NOT see:
```
DEBUG: Fetching user <id>
DEBUG: Fetching user <id>
DEBUG: Fetching user <id>
... (repeated N times)
```

---

## Test 3: Join Requests Optimization

**What changed:** `GetJoinRequestsWithUserInfo()` now uses the `join_requests_with_users` view.

### Test Steps:
1. User A: Create a private room
2. User B: Request to join the room
3. User A: View the room lobby (should see join request with User B's username)

### What to Look For:
- **Server logs** should show:
  ```
  📊 Fetched X join requests with user info (single query via view)
  ```
- **Join request card** should display requester's username correctly
- **Accept/Reject buttons** should work

### Performance Check:
- **Before:** For 10 join requests: 11 queries (1 for requests + 10 for user info)
- **After:** 1 query via the view

---

## Test 4: Full Game Flow (Integration Test)

### Test Steps:
1. **Owner creates room** (tests RoomHandler optimization)
2. **Guest joins room** (tests join flow)
3. **Owner starts game** (tests game state)
4. **Play a few rounds** (tests active game state)
5. **Check `/game/rooms`** (tests ListRoomsHandler optimization)

### What to Look For:
- No errors in server logs (except for that pre-existing room issue)
- All usernames display correctly throughout
- Game state loads correctly
- Performance logs show single-query patterns

---

## Debugging Tips

### If you see duplicate room errors:
```
ERROR: Failed to fetch room <id>: (PGRST116) Cannot coerce the result to a single JSON object
```

This means there might be duplicate rows. Run this query in Supabase:
```sql
SELECT * FROM rooms WHERE id = '<the-error-room-id>';
```

If you see multiple rows with the same ID, you have a serious database integrity issue. Run:
```sql
DELETE FROM rooms WHERE id = '<the-error-room-id>';
```

Then recreate the room fresh.

### If usernames don't appear:
Check the view is working correctly:
```sql
SELECT id, owner_username, guest_username
FROM rooms_with_players
WHERE id = '<your-room-id>';
```

If it returns NULL usernames, check:
```sql
SELECT r.id, r.owner_id, r.guest_id, u1.username as owner_un, u2.username as guest_un
FROM rooms r
LEFT JOIN users u1 ON r.owner_id = u1.id
LEFT JOIN users u2 ON r.guest_id = u2.id
WHERE r.id = '<your-room-id>';
```

---

## Expected Performance Improvements

| Operation | Queries Before | Queries After | Improvement |
|-----------|---------------|---------------|-------------|
| View single room lobby | 3 | 1 | 3x faster |
| List 10 rooms | 12 | 2 | 6x faster |
| View 10 join requests | 11 | 1 | 11x faster |
| View active game state | 5+ | 1 | 5x+ faster |

---

## Success Criteria

✅ All pages load without errors
✅ All usernames display correctly
✅ Server logs show single-query patterns (📊 emoji logs)
✅ No N+1 query patterns in the logs
✅ Game flow works end-to-end
✅ Performance logs confirm optimization

---

## Next Steps After Testing

Once testing is complete:
1. Report any issues found
2. If all tests pass, proceed to Phase 2: SSE HTML Fragments
3. Consider writing automated tests for these optimizations
