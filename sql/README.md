# Database Schema Files

This directory contains the SQL files needed to set up the Couple Card Game database.

## 📦 Quick Start (Fresh Install)

For a brand new installation, run these commands in order:

```bash
# Step 1: Create the complete database schema
psql -U your_username -d your_database -f schema.sql

# Step 2: (Optional) Add sample data for testing
psql -U your_username -d your_database -f seed.sql
```

That's it! 🎉

## 📄 Files

### Essential Files

| File | Purpose | When to Use |
|------|---------|-------------|
| **schema.sql** | Complete database schema with all tables, indexes, triggers, and RLS policies | **Always** - First step of fresh install |
| **seed.sql** | Sample data for testing (categories, questions in 3 languages) | Optional - For development/testing |
| **drop_all.sql** | Drops all tables and extensions (⚠️ DELETES ALL DATA) | When you need to reset the database |

### Documentation

| File | Purpose |
|------|---------|
| **README.md** | This file - installation instructions |

## 🗄️ Database Structure

The `schema.sql` file creates **11 tables**:

### Core Tables
- **users** - User accounts (authenticated and anonymous)
- **friends** - Friend relationships and requests
- **categories** - Question categories (romance, dreams, etc.)
- **questions** - Game questions in multiple languages
- **translations** - UI text translations (i18n)

### Game Tables
- **rooms** - Game rooms where two players play
- **room_join_requests** - Requests to join rooms
- **room_invitations** - Invitations sent by room owners
- **answers** - User answers to questions
- **question_history** - Tracks asked questions per room

### Notification System
- **notifications** - User notifications for events

## ✨ Features Included

The consolidated schema includes:

✅ **Anonymous User Support** - Play without registration  
✅ **Username System** - All users have unique usernames  
✅ **Room Join Requests** - Request to join private rooms  
✅ **Room Invitations** - Invite friends to your room  
✅ **Real-Time Notifications** - Get instant updates  
✅ **Multi-Language Support** - Questions in EN, FR, JA  
✅ **Auto-Updating Timestamps** - Automatic `updated_at` triggers  
✅ **Row Level Security** - Properly configured for anonymous users  

## 🔄 Reset Database

If you need to start fresh (⚠️ **WARNING: DELETES ALL DATA**):

```bash
# Drop all tables
psql -U your_username -d your_database -f drop_all.sql

# Recreate everything
psql -U your_username -d your_database -f schema.sql
psql -U your_username -d your_database -f seed.sql
```

## 🔧 Supabase Setup

If you're using Supabase:

1. Go to your Supabase project dashboard
2. Navigate to **SQL Editor**
3. Copy and paste the contents of `schema.sql`
4. Click **Run**
5. (Optional) Run `seed.sql` for sample data

You should see a success message indicating all tables were created!

## 📊 Schema Details

### Users Table
- Supports both authenticated and anonymous users
- Every user has a unique username (3-20 characters)
- Anonymous users can be promoted to authenticated users

### Rooms Table
Statuses:
- `waiting` - Room created, waiting for guest
- `ready` - Guest joined, ready to play
- `playing` - Game in progress
- `finished` - Game completed

### Notifications
Types:
- `room_invitation` - Someone invited you to a room
- `friend_request` - Someone sent you a friend request
- `game_start` - A game has started
- `message` - Generic message notification

## 🔒 Security

**Row Level Security (RLS):**
- **Disabled** on tables requiring anonymous access (users, rooms, etc.)
- **Enabled** on tables requiring authentication (friends, answers)
- For production, consider using Supabase Service Role keys

**Note:** The current setup prioritizes ease of development. For production:
1. Review and enable RLS on sensitive tables
2. Use service role keys in your backend
3. Implement proper authentication checks
4. Add rate limiting

## 📝 Maintenance

### Adding New Tables
1. Add table definition to `schema.sql` in the appropriate section
2. Add indexes as needed
3. Add to `drop_all.sql` for cleanup
4. Update this README

### Modifying Tables
For existing installations, create migration scripts separately.  
For fresh installs, just update `schema.sql`.

## 🆘 Troubleshooting

### "relation already exists" errors
You're trying to run `schema.sql` on a database that already has tables.  
Solution: Run `drop_all.sql` first, then `schema.sql`

### "permission denied" errors
Make sure you're using a user with sufficient privileges:
```bash
psql -U postgres -d your_database -f schema.sql
```

### Foreign key constraint errors
The schema creates tables in the correct order.  
If you get FK errors, you may have old data. Run `drop_all.sql` first.

## 📚 Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Supabase Documentation](https://supabase.com/docs)
- [Project Documentation](../docs/)

## 🎯 Summary

**Before this consolidation:** 11 SQL files with confusing order  
**After this consolidation:** 3 essential files (+ this README)

Clean, simple, and ready for fresh installs! 🚀
