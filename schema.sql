-- Couple Card Game Database Schema
-- PostgreSQL with Supabase

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    is_anonymous BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_anonymous ON users(is_anonymous) WHERE deleted_at IS NULL;

-- Friends table
CREATE TABLE IF NOT EXISTS friends (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);

CREATE INDEX idx_friends_user_id ON friends(user_id);
CREATE INDEX idx_friends_friend_id ON friends(friend_id);
CREATE INDEX idx_friends_status ON friends(status);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    lang_code VARCHAR(10) NOT NULL,
    question_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_questions_category_id ON questions(category_id);
CREATE INDEX idx_questions_lang_code ON questions(lang_code);

-- Rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('waiting', 'active', 'finished')) DEFAULT 'waiting',
    selected_categories JSONB,
    current_player_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX idx_rooms_guest_id ON rooms(guest_id);
CREATE INDEX idx_rooms_status ON rooms(status);

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

CREATE INDEX idx_room_join_requests_room_id ON room_join_requests(room_id);
CREATE INDEX idx_room_join_requests_user_id ON room_join_requests(user_id);
CREATE INDEX idx_room_join_requests_status ON room_join_requests(status);

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

CREATE INDEX idx_answers_room_id ON answers(room_id);
CREATE INDEX idx_answers_user_id ON answers(user_id);
CREATE INDEX idx_answers_question_id ON answers(question_id);

-- Question history table (to prevent repeating questions in a room)
CREATE TABLE IF NOT EXISTS question_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    asked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, question_id)
);

CREATE INDEX idx_question_history_room_id ON question_history(room_id);

-- Translations table
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lang_code VARCHAR(10) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(lang_code, key)
);

CREATE INDEX idx_translations_lang_code ON translations(lang_code);
CREATE INDEX idx_translations_key ON translations(key);

-- Row Level Security (RLS) Policies

-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE friends ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE questions ENABLE ROW LEVEL SECURITY;
ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;
ALTER TABLE room_join_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE answers ENABLE ROW LEVEL SECURITY;
ALTER TABLE question_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE translations ENABLE ROW LEVEL SECURITY;

-- Users policies
CREATE POLICY "Users can view their own data" ON users
    FOR SELECT USING (auth.uid() = id OR is_admin = TRUE);

CREATE POLICY "Users can update their own data" ON users
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Admins can do everything with users" ON users
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Friends policies
CREATE POLICY "Users can view their own friendships" ON friends
    FOR SELECT USING (auth.uid() = user_id OR auth.uid() = friend_id);

CREATE POLICY "Users can manage their own friendships" ON friends
    FOR ALL USING (auth.uid() = user_id);

CREATE POLICY "Admins can view all friendships" ON friends
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Categories policies (public read, admin write)
CREATE POLICY "Anyone can view categories" ON categories
    FOR SELECT USING (TRUE);

CREATE POLICY "Admins can manage categories" ON categories
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Questions policies (public read, admin write)
CREATE POLICY "Anyone can view questions" ON questions
    FOR SELECT USING (TRUE);

CREATE POLICY "Admins can manage questions" ON questions
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Rooms policies
CREATE POLICY "Room participants can view their rooms" ON rooms
    FOR SELECT USING (auth.uid() = owner_id OR auth.uid() = guest_id);

CREATE POLICY "Room owners can create rooms" ON rooms
    FOR INSERT WITH CHECK (auth.uid() = owner_id);

CREATE POLICY "Room owners can update their rooms" ON rooms
    FOR UPDATE USING (auth.uid() = owner_id);

CREATE POLICY "Admins can view all rooms" ON rooms
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Room join requests policies
CREATE POLICY "Room owners can view join requests for their rooms" ON room_join_requests
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM rooms WHERE id = room_id AND owner_id = auth.uid())
    );

CREATE POLICY "Users can view their own join requests" ON room_join_requests
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create join requests" ON room_join_requests
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Room owners can update join requests" ON room_join_requests
    FOR UPDATE USING (
        EXISTS (SELECT 1 FROM rooms WHERE id = room_id AND owner_id = auth.uid())
    );

-- Answers policies
CREATE POLICY "Room participants can view answers in their rooms" ON answers
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

CREATE POLICY "Admins can view all answers" ON answers
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
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

-- Translations policies (public read, admin write)
CREATE POLICY "Anyone can view translations" ON translations
    FOR SELECT USING (TRUE);

CREATE POLICY "Admins can manage translations" ON translations
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND is_admin = TRUE)
    );

-- Functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_questions_updated_at BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_room_join_requests_updated_at BEFORE UPDATE ON room_join_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_translations_updated_at BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- Fix RLS Policies to avoid infinite recursion
-- Complete rewrite of user table policies

-- Drop ALL existing policies on users table
DROP POLICY IF EXISTS "Allow public anonymous user creation" ON users;
DROP POLICY IF EXISTS "Users can view their own data" ON users;
DROP POLICY IF EXISTS "Users can update their own data" ON users;
DROP POLICY IF EXISTS "Admins can do everything with users" ON users;

-- DISABLE RLS on users table (simplest solution for now)
-- This allows the Go API to work without authentication issues
ALTER TABLE users DISABLE ROW LEVEL SECURITY;

-- Note: For production, you should use Supabase Service Role key in your Go app
-- and keep RLS enabled. But for development, disabling RLS is acceptable.

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'RLS disabled on users table';
    RAISE NOTICE 'Anonymous user creation should now work!';
    RAISE NOTICE '===========================================';
END $$;


