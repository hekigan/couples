-- Couple Card Game Database Schema
-- PostgreSQL / Supabase

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For text search

-- Create update_updated_at_column function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================================================
-- Users Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    username VARCHAR(50) UNIQUE,
    display_name VARCHAR(100),
    avatar_url TEXT,
    is_anonymous BOOLEAN DEFAULT FALSE,
    language_preference VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_anonymous ON users(is_anonymous);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Categories Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(50) UNIQUE NOT NULL,
    icon VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_key ON categories(key);

ALTER TABLE categories ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Categories are viewable by everyone"
ON categories FOR SELECT
TO public
USING (true);

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Questions Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL DEFAULT 'en',
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questions_category_id ON questions(category_id);
CREATE INDEX IF NOT EXISTS idx_questions_language_code ON questions(language_code);
CREATE INDEX IF NOT EXISTS idx_questions_text_trgm ON questions USING gin (text gin_trgm_ops);

ALTER TABLE questions ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Questions are viewable by everyone"
ON questions FOR SELECT
TO public
USING (true);

CREATE TRIGGER update_questions_updated_at BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Rooms Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'ready', 'playing', 'finished')),
    language VARCHAR(10) DEFAULT 'en',
    is_private BOOLEAN DEFAULT FALSE,
    max_questions INTEGER DEFAULT 10,
    current_question INTEGER DEFAULT 0,
    current_turn UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_guest_id ON rooms(guest_id);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);
CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at);

ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own rooms"
ON rooms FOR SELECT
TO public
USING (
    owner_id = auth.uid() OR 
    guest_id = auth.uid() OR
    is_private = FALSE OR
    EXISTS (SELECT 1 FROM users WHERE users.id = owner_id AND users.is_anonymous = TRUE) OR
    EXISTS (SELECT 1 FROM users WHERE users.id = guest_id AND users.is_anonymous = TRUE)
);

CREATE POLICY "Users can create rooms"
ON rooms FOR INSERT
TO public
WITH CHECK (owner_id = auth.uid() OR EXISTS (
    SELECT 1 FROM users WHERE users.id = owner_id AND users.is_anonymous = TRUE
));

CREATE POLICY "Room participants can update"
ON rooms FOR UPDATE
TO public
USING (owner_id = auth.uid() OR guest_id = auth.uid());

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Question History Table (tracks questions asked in each room)
-- ============================================================================
CREATE TABLE IF NOT EXISTS question_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    asked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_question_history_room_id ON question_history(room_id);
CREATE INDEX IF NOT EXISTS idx_question_history_question_id ON question_history(question_id);

ALTER TABLE question_history ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Room participants can view question history"
ON question_history FOR SELECT
TO public
USING (
    room_id IN (
        SELECT id FROM rooms WHERE owner_id = auth.uid() OR guest_id = auth.uid()
    )
);

CREATE POLICY "Room participants can add to question history"
ON question_history FOR INSERT
TO public
WITH CHECK (
    room_id IN (
        SELECT id FROM rooms WHERE owner_id = auth.uid() OR guest_id = auth.uid()
    )
);

-- ============================================================================
-- Answers Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    answer_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_answers_room_id ON answers(room_id);
CREATE INDEX IF NOT EXISTS idx_answers_question_id ON answers(question_id);
CREATE INDEX IF NOT EXISTS idx_answers_user_id ON answers(user_id);

ALTER TABLE answers ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Room participants can view answers"
ON answers FOR SELECT
TO public
USING (
    room_id IN (
        SELECT id FROM rooms WHERE owner_id = auth.uid() OR guest_id = auth.uid()
    )
);

CREATE POLICY "Users can create their own answers"
ON answers FOR INSERT
TO public
WITH CHECK (user_id = auth.uid() OR EXISTS (
    SELECT 1 FROM users WHERE users.id = user_id AND users.is_anonymous = TRUE
));

-- ============================================================================
-- Friends Table (bidirectional friendship)
-- ============================================================================
CREATE TABLE IF NOT EXISTS friends (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, friend_id),
    CHECK (user_id != friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friends_user_id ON friends(user_id);
CREATE INDEX IF NOT EXISTS idx_friends_friend_id ON friends(friend_id);

ALTER TABLE friends ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own friendships"
ON friends FOR SELECT
TO public
USING (user_id = auth.uid() OR friend_id = auth.uid());

CREATE POLICY "Users can create friendships"
ON friends FOR INSERT
TO public
WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can delete friendships"
ON friends FOR DELETE
TO public
USING (user_id = auth.uid() OR friend_id = auth.uid());

-- ============================================================================
-- Friend Requests Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS friend_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(sender_id, receiver_id),
    CHECK (sender_id != receiver_id)
);

CREATE INDEX IF NOT EXISTS idx_friend_requests_sender_id ON friend_requests(sender_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver_id ON friend_requests(receiver_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_status ON friend_requests(status);

ALTER TABLE friend_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own friend requests"
ON friend_requests FOR SELECT
TO public
USING (sender_id = auth.uid() OR receiver_id = auth.uid());

CREATE POLICY "Users can create friend requests"
ON friend_requests FOR INSERT
TO public
WITH CHECK (sender_id = auth.uid());

CREATE POLICY "Receivers can update friend requests"
ON friend_requests FOR UPDATE
TO public
USING (receiver_id = auth.uid());

CREATE TRIGGER update_friend_requests_updated_at BEFORE UPDATE ON friend_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Translations Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    language_code VARCHAR(10) NOT NULL,
    translation_key VARCHAR(100) NOT NULL,
    translation_value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(language_code, translation_key)
);

CREATE INDEX IF NOT EXISTS idx_translations_language_code ON translations(language_code);
CREATE INDEX IF NOT EXISTS idx_translations_key ON translations(translation_key);

ALTER TABLE translations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Translations are viewable by everyone"
ON translations FOR SELECT
TO public
USING (true);

CREATE TRIGGER update_translations_updated_at BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Room Join Requests Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS room_join_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, user_id, status)
);

CREATE INDEX IF NOT EXISTS idx_room_join_requests_room_id ON room_join_requests(room_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_user_id ON room_join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_status ON room_join_requests(status);

ALTER TABLE room_join_requests ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_room_join_requests_updated_at BEFORE UPDATE ON room_join_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Database schema created successfully!';
    RAISE NOTICE 'Next step: Run seed.sql for sample data';
    RAISE NOTICE '===========================================';
END $$;

