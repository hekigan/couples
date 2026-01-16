-- ============================================================================
-- Couple Card Game Database Schema
-- PostgreSQL with Supabase
-- Last Updated: January 17, 2026
-- ============================================================================
--
-- SECURITY NOTES:
-- 1. uuid-ossp extension is in the 'extensions' schema (best practice)
-- 2. All tables have Row Level Security (RLS) enabled
-- 3. Functions use SET search_path = public to prevent search path attacks
-- 4. The Go backend uses SERVICE_ROLE_KEY which bypasses RLS
-- 5. RLS policies protect against direct API access via anon key
-- ============================================================================

-- Create extensions schema for extension isolation (security best practice)
CREATE SCHEMA IF NOT EXISTS extensions;

-- Enable UUID extension in extensions schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA extensions;

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    username VARCHAR(50) UNIQUE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    is_anonymous BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_anonymous ON users(is_anonymous) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

COMMENT ON TABLE users IS 'Application users - supports both authenticated and anonymous users';

-- Friends table
CREATE TABLE IF NOT EXISTS friends (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friends_user_id ON friends(user_id);
CREATE INDEX IF NOT EXISTS idx_friends_friend_id ON friends(friend_id);
CREATE INDEX IF NOT EXISTS idx_friends_status ON friends(status);

COMMENT ON TABLE friends IS 'User friendships and friend requests';

-- Friend email invitations table
CREATE TABLE IF NOT EXISTS friend_email_invitations (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_email VARCHAR(255) NOT NULL,
    token VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'accepted', 'expired', 'cancelled')) DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_sender_recipient UNIQUE(sender_id, recipient_email)
);

CREATE INDEX IF NOT EXISTS idx_friend_email_invitations_sender ON friend_email_invitations(sender_id);
CREATE INDEX IF NOT EXISTS idx_friend_email_invitations_token ON friend_email_invitations(token);
CREATE INDEX IF NOT EXISTS idx_friend_email_invitations_email ON friend_email_invitations(recipient_email);
CREATE INDEX IF NOT EXISTS idx_friend_email_invitations_status ON friend_email_invitations(status);
CREATE INDEX IF NOT EXISTS idx_friend_email_invitations_expires ON friend_email_invitations(expires_at);

COMMENT ON TABLE friend_email_invitations IS 'Email-based friend invitations for users not yet in the system';
COMMENT ON COLUMN friend_email_invitations.token IS 'Secure token for email invitation acceptance link';
COMMENT ON COLUMN friend_email_invitations.expires_at IS 'Invitation expiration timestamp (7 days default)';

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    label VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_key ON categories(key);

COMMENT ON TABLE categories IS 'Question categories (e.g., romance, dreams, past)';
COMMENT ON COLUMN categories.key IS 'Internal key for category (e.g., couples, friends, sex)';
COMMENT ON COLUMN categories.label IS 'Human-readable display name for the category';

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    lang_code VARCHAR(10) NOT NULL,
    question_text TEXT NOT NULL,
    base_question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questions_category_id ON questions(category_id);
CREATE INDEX IF NOT EXISTS idx_questions_lang_code ON questions(lang_code);
CREATE INDEX IF NOT EXISTS idx_questions_category_lang ON questions(category_id, lang_code);
CREATE INDEX IF NOT EXISTS idx_questions_base_question_id ON questions(base_question_id);

COMMENT ON TABLE questions IS 'Game questions in multiple languages';
COMMENT ON COLUMN questions.base_question_id IS 'Links translations together. English questions reference themselves, translations reference the English version.';

-- Rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    name VARCHAR(255),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES users(id) ON DELETE CASCADE,
    -- IMPORTANT: 'ready' status is REQUIRED for join request flow!
    -- When a guest joins via join request, status changes: waiting → ready → playing
    status VARCHAR(50) NOT NULL CHECK (status IN ('waiting', 'ready', 'playing', 'finished')) DEFAULT 'waiting',
    language VARCHAR(10) DEFAULT 'en',
    is_private BOOLEAN DEFAULT FALSE,
    guest_ready BOOLEAN DEFAULT FALSE,
    max_questions INT DEFAULT 20,
    current_question INT DEFAULT 0,
    current_question_id UUID REFERENCES questions(id),
    selected_categories JSONB,
    current_player_id UUID REFERENCES users(id),
    paused_at TIMESTAMP WITH TIME ZONE,
    disconnected_user UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_guest_id ON rooms(guest_id);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);
CREATE INDEX IF NOT EXISTS idx_rooms_language ON rooms(language);
CREATE INDEX IF NOT EXISTS idx_rooms_current_question_id ON rooms(current_question_id);
CREATE INDEX IF NOT EXISTS idx_rooms_disconnected_user ON rooms(disconnected_user);

COMMENT ON TABLE rooms IS 'Game rooms where two players play together';
COMMENT ON COLUMN rooms.name IS 'Optional room name set by owner';
COMMENT ON COLUMN rooms.status IS 'waiting=no guest, ready=guest joined, playing=game active, finished=game over';
COMMENT ON COLUMN rooms.language IS 'Game language (en, fr, ja, etc.)';
COMMENT ON COLUMN rooms.is_private IS 'Whether room requires invitation to join';
COMMENT ON COLUMN rooms.max_questions IS 'Maximum number of questions for this game';
COMMENT ON COLUMN rooms.current_question IS 'Current question number (0-based)';
COMMENT ON COLUMN rooms.current_question_id IS 'ID of the currently active question (persists across page refreshes)';
COMMENT ON COLUMN rooms.paused_at IS 'Timestamp when game was paused (if paused)';
COMMENT ON COLUMN rooms.disconnected_user IS 'User who disconnected (if any)';

-- Room join requests table
CREATE TABLE IF NOT EXISTS room_join_requests (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected')) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_room_join_requests_room_id ON room_join_requests(room_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_user_id ON room_join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_status ON room_join_requests(status);

COMMENT ON TABLE room_join_requests IS 'Requests from users to join private rooms';

-- Room invitations table
CREATE TABLE IF NOT EXISTS room_invitations (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, invitee_id)
);

CREATE INDEX IF NOT EXISTS idx_room_invitations_invitee ON room_invitations(invitee_id);
CREATE INDEX IF NOT EXISTS idx_room_invitations_room ON room_invitations(room_id);
CREATE INDEX IF NOT EXISTS idx_room_invitations_status ON room_invitations(status);

COMMENT ON TABLE room_invitations IS 'Room owner invitations sent to friends';

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    link VARCHAR(500),
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, read);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

COMMENT ON TABLE notifications IS 'User notifications for invitations, friend requests, etc.';
COMMENT ON COLUMN notifications.type IS 'room_invitation, friend_request, game_start, message';

-- Answers table
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    answer_text TEXT,
    action_type VARCHAR(50) NOT NULL CHECK (action_type IN ('answered', 'skipped')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_answers_room_id ON answers(room_id);
CREATE INDEX IF NOT EXISTS idx_answers_user_id ON answers(user_id);
CREATE INDEX IF NOT EXISTS idx_answers_question_id ON answers(question_id);

COMMENT ON TABLE answers IS 'User answers or skips to game questions';

-- Question history table (prevents repeating questions in a room)
CREATE TABLE IF NOT EXISTS question_history (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    asked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_question_history_room_id ON question_history(room_id);

COMMENT ON TABLE question_history IS 'Tracks which questions have been asked in each room';

-- Translations table
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
    lang_code VARCHAR(10) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(lang_code, key)
);

CREATE INDEX IF NOT EXISTS idx_translations_lang_code ON translations(lang_code);
CREATE INDEX IF NOT EXISTS idx_translations_key ON translations(key);

COMMENT ON TABLE translations IS 'UI translations for internationalization (i18n)';

-- ============================================================================
-- FUNCTIONS AND TRIGGERS
-- ============================================================================

-- Function for automatic timestamp updates
-- SET search_path prevents search path manipulation attacks
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql
SET search_path = public;

-- Triggers for updated_at columns
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_questions_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rooms_updated_at
    BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_room_join_requests_updated_at
    BEFORE UPDATE ON room_join_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_room_invitations_updated_at
    BEFORE UPDATE ON room_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_friend_email_invitations_updated_at
    BEFORE UPDATE ON friend_email_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_translations_updated_at
    BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================================
--
-- IMPORTANT: The Go backend uses SUPABASE_SERVICE_ROLE_KEY which bypasses RLS.
-- These policies protect against direct PostgREST API access via anon key.
--
-- Strategy:
-- - Public read tables (categories, questions, translations): SELECT for all
-- - User-owned data (notifications): Full access for owner
-- - Participant data (rooms, requests, invitations): Access for participants
-- - Sensitive data (friend_email_invitations): Restricted by sender, token lookup allowed
-- ============================================================================

-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;
ALTER TABLE room_join_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE room_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE friend_email_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE friends ENABLE ROW LEVEL SECURITY;
ALTER TABLE answers ENABLE ROW LEVEL SECURITY;
ALTER TABLE question_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE questions ENABLE ROW LEVEL SECURITY;
ALTER TABLE translations ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- PUBLIC READ-ONLY TABLES POLICIES
-- Categories, Questions, and Translations are public read-only
-- ============================================================================

-- Categories: Public read access
CREATE POLICY "categories_select_all" ON categories
    FOR SELECT USING (true);

-- Questions: Public read access
CREATE POLICY "questions_select_all" ON questions
    FOR SELECT USING (true);

-- Translations: Public read access
CREATE POLICY "translations_select_all" ON translations
    FOR SELECT USING (true);

-- ============================================================================
-- USER TABLE POLICIES
-- Users can view other users (for friend search, room display)
-- Users can only update their own profile
-- ============================================================================

-- Users: Public read access for non-deleted users
CREATE POLICY "users_select_active" ON users
    FOR SELECT USING (deleted_at IS NULL);

-- Users: Self-update only
CREATE POLICY "users_update_self" ON users
    FOR UPDATE USING (auth.uid() = id);

-- ============================================================================
-- ROOM POLICIES
-- Room participants (owner or guest) can view and manage their rooms
-- Public rooms are visible to all for joining
-- ============================================================================

-- Rooms: View public rooms or rooms you participate in
CREATE POLICY "rooms_select_public_or_participant" ON rooms
    FOR SELECT USING (
        is_private = false
        OR owner_id = auth.uid()
        OR guest_id = auth.uid()
    );

-- Rooms: Owner can update their rooms
CREATE POLICY "rooms_update_owner" ON rooms
    FOR UPDATE USING (owner_id = auth.uid());

-- Rooms: Owner can delete their rooms
CREATE POLICY "rooms_delete_owner" ON rooms
    FOR DELETE USING (owner_id = auth.uid());

-- Rooms: Authenticated users can create rooms
CREATE POLICY "rooms_insert_authenticated" ON rooms
    FOR INSERT WITH CHECK (auth.uid() = owner_id);

-- ============================================================================
-- ROOM JOIN REQUESTS POLICIES
-- Users can view requests for rooms they own or requests they made
-- ============================================================================

-- Join requests: View own requests or requests to rooms you own
CREATE POLICY "room_join_requests_select" ON room_join_requests
    FOR SELECT USING (
        user_id = auth.uid()
        OR EXISTS (
            SELECT 1 FROM rooms WHERE id = room_id AND owner_id = auth.uid()
        )
    );

-- Join requests: Users can create requests
CREATE POLICY "room_join_requests_insert" ON room_join_requests
    FOR INSERT WITH CHECK (user_id = auth.uid());

-- Join requests: Room owner can update request status
CREATE POLICY "room_join_requests_update_owner" ON room_join_requests
    FOR UPDATE USING (
        EXISTS (
            SELECT 1 FROM rooms WHERE id = room_id AND owner_id = auth.uid()
        )
    );

-- Join requests: Requester can delete their own request
CREATE POLICY "room_join_requests_delete_self" ON room_join_requests
    FOR DELETE USING (user_id = auth.uid());

-- ============================================================================
-- ROOM INVITATIONS POLICIES
-- Inviter or invitee can view/manage invitations
-- ============================================================================

-- Invitations: Inviter or invitee can view
CREATE POLICY "room_invitations_select" ON room_invitations
    FOR SELECT USING (
        inviter_id = auth.uid()
        OR invitee_id = auth.uid()
    );

-- Invitations: Room owner (inviter) can create
CREATE POLICY "room_invitations_insert" ON room_invitations
    FOR INSERT WITH CHECK (inviter_id = auth.uid());

-- Invitations: Invitee can update (accept/decline)
CREATE POLICY "room_invitations_update_invitee" ON room_invitations
    FOR UPDATE USING (invitee_id = auth.uid());

-- Invitations: Inviter can cancel (delete)
CREATE POLICY "room_invitations_delete_inviter" ON room_invitations
    FOR DELETE USING (inviter_id = auth.uid());

-- ============================================================================
-- NOTIFICATIONS POLICIES
-- Users can only access their own notifications
-- ============================================================================

-- Notifications: User can view own notifications
CREATE POLICY "notifications_select_own" ON notifications
    FOR SELECT USING (user_id = auth.uid());

-- Notifications: User can update own notifications (mark as read)
CREATE POLICY "notifications_update_own" ON notifications
    FOR UPDATE USING (user_id = auth.uid());

-- Notifications: User can delete own notifications
CREATE POLICY "notifications_delete_own" ON notifications
    FOR DELETE USING (user_id = auth.uid());

-- Notifications: Users can only insert notifications for themselves
-- Note: Backend handles all notification creation with service role key
CREATE POLICY "notifications_insert" ON notifications
    FOR INSERT WITH CHECK (user_id = auth.uid());

-- ============================================================================
-- FRIEND EMAIL INVITATIONS POLICIES
-- Sender can manage their invitations
-- Anyone can look up by token (for acceptance)
-- ============================================================================

-- Friend email invitations: Sender can view their invitations
CREATE POLICY "friend_email_invitations_select_sender" ON friend_email_invitations
    FOR SELECT USING (sender_id = auth.uid());

-- Friend email invitations: Allow token lookup for acceptance (read-only)
-- This is intentionally permissive for the token-based acceptance flow
CREATE POLICY "friend_email_invitations_select_by_token" ON friend_email_invitations
    FOR SELECT USING (true);

-- Friend email invitations: Sender can create
CREATE POLICY "friend_email_invitations_insert" ON friend_email_invitations
    FOR INSERT WITH CHECK (sender_id = auth.uid());

-- Friend email invitations: Only sender can update (cancel)
-- Note: Token-based acceptance is handled by backend with service role key
CREATE POLICY "friend_email_invitations_update" ON friend_email_invitations
    FOR UPDATE USING (sender_id = auth.uid());

-- Friend email invitations: Sender can delete
CREATE POLICY "friend_email_invitations_delete" ON friend_email_invitations
    FOR DELETE USING (sender_id = auth.uid());

-- ============================================================================
-- FRIENDS TABLE POLICIES
-- Users can view their own friendships
-- Users can manage friendships they initiated
-- ============================================================================

-- Friends: View friendships you're part of
CREATE POLICY "friends_select_own" ON friends
    FOR SELECT USING (
        user_id = auth.uid()
        OR friend_id = auth.uid()
    );

-- Friends: Create friend request
CREATE POLICY "friends_insert" ON friends
    FOR INSERT WITH CHECK (user_id = auth.uid());

-- Friends: Update friendship (accept/decline) - both parties
CREATE POLICY "friends_update" ON friends
    FOR UPDATE USING (
        user_id = auth.uid()
        OR friend_id = auth.uid()
    );

-- Friends: Delete friendship - both parties
CREATE POLICY "friends_delete" ON friends
    FOR DELETE USING (
        user_id = auth.uid()
        OR friend_id = auth.uid()
    );

-- ============================================================================
-- ANSWERS TABLE POLICIES
-- Room participants can view and create answers
-- ============================================================================

-- Answers: Room participants can view
CREATE POLICY "answers_select_participant" ON answers
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM rooms
            WHERE id = room_id
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

-- Answers: Room participants can create
CREATE POLICY "answers_insert_participant" ON answers
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM rooms
            WHERE id = room_id
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

-- ============================================================================
-- QUESTION HISTORY TABLE POLICIES
-- Room participants can view and add to question history
-- ============================================================================

-- Question history: Room participants can view
CREATE POLICY "question_history_select_participant" ON question_history
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM rooms
            WHERE id = room_id
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

-- Question history: Room participants can add
CREATE POLICY "question_history_insert_participant" ON question_history
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM rooms
            WHERE id = room_id
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

-- ============================================================================
-- COMPLETION MESSAGE
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '============================================================';
    RAISE NOTICE 'Couple Card Game Database Schema - Installation Complete';
    RAISE NOTICE '============================================================';
    RAISE NOTICE '';
    RAISE NOTICE 'Tables Created:';
    RAISE NOTICE '  ✓ users (with username support)';
    RAISE NOTICE '  ✓ friends';
    RAISE NOTICE '  ✓ friend_email_invitations';
    RAISE NOTICE '  ✓ categories';
    RAISE NOTICE '  ✓ questions (multi-language)';
    RAISE NOTICE '  ✓ rooms';
    RAISE NOTICE '  ✓ room_join_requests';
    RAISE NOTICE '  ✓ room_invitations';
    RAISE NOTICE '  ✓ notifications';
    RAISE NOTICE '  ✓ answers';
    RAISE NOTICE '  ✓ question_history';
    RAISE NOTICE '  ✓ translations';
    RAISE NOTICE '';
    RAISE NOTICE 'Security Features Enabled:';
    RAISE NOTICE '  ✓ Row Level Security on all tables';
    RAISE NOTICE '  ✓ uuid-ossp extension in extensions schema';
    RAISE NOTICE '  ✓ Functions with immutable search_path';
    RAISE NOTICE '';
    RAISE NOTICE 'Features Enabled:';
    RAISE NOTICE '  ✓ Anonymous user support';
    RAISE NOTICE '  ✓ Room join requests system';
    RAISE NOTICE '  ✓ Room invitation system';
    RAISE NOTICE '  ✓ Real-time notifications';
    RAISE NOTICE '  ✓ Multi-language support';
    RAISE NOTICE '  ✓ Auto-updating timestamps';
    RAISE NOTICE '';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '  1. Run seed.sql to populate with sample data';
    RAISE NOTICE '  2. Configure .env with database credentials';
    RAISE NOTICE '  3. Start the Go server: ./couple-game';
    RAISE NOTICE '';
    RAISE NOTICE '============================================================';
END $$;
