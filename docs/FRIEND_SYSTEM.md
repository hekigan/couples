# 👥 Friend System Guide

Complete guide to the friend invitation and management system.

## Overview

The friend system allows users to:
- Send friend invitations by UUID
- Accept or decline invitations
- View friends list
- Play games with friends
- Remove friendships

## Features

### 1. Friend Invitations

**Sending Invitations:**

1. Navigate to `/friends`
2. Click "Add Friend"
3. Enter friend's User UUID
4. Click "Send Invitation"

**Finding Your UUID:**

Users can find their UUID in their profile or share it with friends.

### 2. Managing Invitations

**Accepting:**

1. Go to `/friends`
2. See pending invitations section
3. Click "Accept" on an invitation
4. Friend is added to your list

**Declining:**

1. Go to `/friends`
2. Click "Decline" on an invitation
3. Invitation is removed

### 3. Friends List

View all your friends at `/friends`:

- Friend's display name
- Email (if available)
- Quick actions:
  - Play together
  - Remove friend

### 4. Playing with Friends

1. Create a room
2. Friend will see invitation
3. Click join to play together

## Database Schema

### Friends Table

```sql
CREATE TABLE friends (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    friend_id UUID REFERENCES users(id),
    created_at TIMESTAMP,
    UNIQUE(user_id, friend_id)
);
```

### Friend Requests Table

```sql
CREATE TABLE friend_requests (
    id UUID PRIMARY KEY,
    sender_id UUID REFERENCES users(id),
    receiver_id UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## API Endpoints

### Get Friends List

```
GET /api/friends
Authorization: Required
Response: Array of friend objects
```

### Send Friend Request

```
POST /api/friends/add
Authorization: Required
Body: { "friend_identifier": "uuid" }
Response: { "success": true }
```

### Accept Friend Request

```
POST /api/friends/accept/{id}
Authorization: Required
Response: { "success": true }
```

### Decline Friend Request

```
POST /api/friends/decline/{id}
Authorization: Required
Response: { "success": true }
```

### Remove Friend

```
DELETE /api/friends/remove/{id}
Authorization: Required
Response: { "success": true }
```

## UI Components

### Friends Page (`/friends`)

Displays:
- Active friends list
- Pending received invitations
- Pending sent invitations
- Add friend button

### Add Friend Modal

Form to send invitation:
- Friend UUID input
- Send button
- Cancel button

### Friend Card Component

Shows:
- Friend avatar (if available)
- Friend display name
- Friend email/identifier
- Action buttons:
  - Play together
  - Remove friendship

## HTMX Integration

The friend system uses HTMX for dynamic updates:

```html
<!-- Accept invitation -->
<button hx-post="/api/friends/accept/{id}"
        hx-target="#friend-requests"
        hx-swap="outerHTML">
  Accept
</button>

<!-- Decline invitation -->
<button hx-post="/api/friends/decline/{id}"
        hx-target="#friend-requests"
        hx-swap="outerHTML">
  Decline
</button>
```

## Security

### Row Level Security (RLS)

Friends table policies:

```sql
-- Users can view their own friendships
CREATE POLICY "Users can view their friendships"
ON friends FOR SELECT
USING (user_id = auth.uid() OR friend_id = auth.uid());

-- Users can create friendships
CREATE POLICY "Users can create friendships"
ON friends FOR INSERT
WITH CHECK (user_id = auth.uid());

-- Users can delete friendships
CREATE POLICY "Users can delete friendships"
ON friends FOR DELETE
USING (user_id = auth.uid() OR friend_id = auth.uid());
```

Friend requests policies:

```sql
-- Users can view requests they're involved in
CREATE POLICY "Users can view their requests"
ON friend_requests FOR SELECT
USING (sender_id = auth.uid() OR receiver_id = auth.uid());

-- Users can create requests
CREATE POLICY "Users can create requests"
ON friend_requests FOR INSERT
WITH CHECK (sender_id = auth.uid());

-- Receivers can update requests
CREATE POLICY "Receivers can update requests"
ON friend_requests FOR UPDATE
USING (receiver_id = auth.uid());
```

## Customization

### Styling

Friend components use SASS styles in `sass/pages/_friends.scss`:

```scss
.friend-card {
  background-color: $bg-primary;
  border-radius: $border-radius-lg;
  padding: $spacing-lg;
  // ... custom styles
}
```

### Templates

Friend templates in `templates/friends/`:
- `list.html` - Main friends page
- `add.html` - Add friend form (if separate)

### Translations

Add translations in `static/i18n/{language}.json`:

```json
{
  "friends.title": "Friends",
  "friends.add": "Add Friend",
  "friends.pending": "Pending Invitations",
  "friends.none": "No friends yet"
}
```

## Troubleshooting

### Can't Find Friend UUID

- Users need to share their UUID
- UUID shown in profile page
- Can implement username search (future feature)

### Invitation Not Received

- Check sender/receiver UUIDs correct
- Verify database insert succeeded
- Check RLS policies allow access

### Can't Remove Friend

- Verify friendship exists
- Check RLS policies
- Ensure user owns friendship

## Future Enhancements

Potential additions:
- Search friends by username/email
- Friend suggestions
- Block users
- Friend activity feed
- Group chats
- Multiplayer games

---

**Friend System Complete!** 👥 Start building your circle!

