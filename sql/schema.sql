-- ============================================================================
-- Couple Card Game Database Schema
-- PostgreSQL with Supabase
-- Last Updated: November 6, 2025
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_key ON categories(key);

COMMENT ON TABLE categories IS 'Question categories (e.g., romance, dreams, past)';

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    lang_code VARCHAR(10) NOT NULL,
    question_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questions_category_id ON questions(category_id);
CREATE INDEX IF NOT EXISTS idx_questions_lang_code ON questions(lang_code);
CREATE INDEX IF NOT EXISTS idx_questions_category_lang ON questions(category_id, lang_code);

COMMENT ON TABLE questions IS 'Game questions in multiple languages';

-- Rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES users(id) ON DELETE CASCADE,
    -- IMPORTANT: 'ready' status is REQUIRED for join request flow!
    -- When a guest joins via join request, status changes: waiting → ready → playing
    status VARCHAR(50) NOT NULL CHECK (status IN ('waiting', 'ready', 'playing', 'finished')) DEFAULT 'waiting',
    selected_categories JSONB,
    current_player_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_guest_id ON rooms(guest_id);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);

COMMENT ON TABLE rooms IS 'Game rooms where two players play together';
COMMENT ON COLUMN rooms.status IS 'waiting=no guest, ready=guest joined, playing=game active, finished=game over';

-- Room join requests table
CREATE TABLE IF NOT EXISTS room_join_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    answer_text TEXT,
    action_type VARCHAR(50) NOT NULL CHECK (action_type IN ('answered', 'passed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_answers_room_id ON answers(room_id);
CREATE INDEX IF NOT EXISTS idx_answers_user_id ON answers(user_id);
CREATE INDEX IF NOT EXISTS idx_answers_question_id ON answers(question_id);

COMMENT ON TABLE answers IS 'User answers or passes to game questions';

-- Question history table (prevents repeating questions in a room)
CREATE TABLE IF NOT EXISTS question_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    asked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_question_history_room_id ON question_history(room_id);

COMMENT ON TABLE question_history IS 'Tracks which questions have been asked in each room';

-- Translations table
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

CREATE TRIGGER update_translations_updated_at 
    BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================================

-- Note: RLS is DISABLED for development to support anonymous users.
-- The Go backend handles authentication and authorization.
--
-- For production deployment:
-- 1. Use Supabase Service Role key in your Go app
-- 2. Enable RLS on sensitive tables
-- 3. Create appropriate policies for authenticated users
-- 4. Keep public tables (categories, questions, translations) open for read

-- Disable RLS on tables that need anonymous user access
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE rooms DISABLE ROW LEVEL SECURITY;
ALTER TABLE room_join_requests DISABLE ROW LEVEL SECURITY;
ALTER TABLE room_invitations DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;

-- Enable RLS on tables with appropriate policies
ALTER TABLE friends ENABLE ROW LEVEL SECURITY;
ALTER TABLE answers ENABLE ROW LEVEL SECURITY;
ALTER TABLE question_history ENABLE ROW LEVEL SECURITY;

-- Public read-only tables (no RLS needed)
ALTER TABLE categories DISABLE ROW LEVEL SECURITY;
ALTER TABLE questions DISABLE ROW LEVEL SECURITY;
ALTER TABLE translations DISABLE ROW LEVEL SECURITY;

-- Friends policies
CREATE POLICY "Users can view their own friendships" ON friends
    FOR SELECT USING (auth.uid() = user_id OR auth.uid() = friend_id);

CREATE POLICY "Users can manage their own friendships" ON friends
    FOR ALL USING (auth.uid() = user_id);

-- Answers policies (only for authenticated users)
CREATE POLICY "Room participants can view answers" ON answers
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM rooms 
            WHERE id = room_id 
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

CREATE POLICY "Room participants can create answers" ON answers
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM rooms 
            WHERE id = room_id 
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

-- Question history policies
CREATE POLICY "Room participants can view question history" ON question_history
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM rooms 
            WHERE id = room_id 
            AND (owner_id = auth.uid() OR guest_id = auth.uid())
        )
    );

CREATE POLICY "Room participants can add to question history" ON question_history
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
