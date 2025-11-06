-- Migration V2: Add room join requests functionality
-- Safe to run multiple times (uses IF NOT EXISTS)

-- Create room_join_requests table
CREATE TABLE IF NOT EXISTS room_join_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, user_id, status) -- Prevent duplicate pending requests
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_room_join_requests_room_id ON room_join_requests(room_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_user_id ON room_join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_status ON room_join_requests(status);
CREATE INDEX IF NOT EXISTS idx_room_join_requests_created_at ON room_join_requests(created_at);

-- Enable Row Level Security
ALTER TABLE room_join_requests ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS "Users can create join requests" ON room_join_requests;
DROP POLICY IF EXISTS "Users can view their own requests" ON room_join_requests;
DROP POLICY IF EXISTS "Room owners can view requests for their rooms" ON room_join_requests;
DROP POLICY IF EXISTS "Room owners can update requests for their rooms" ON room_join_requests;

-- Create RLS policies
CREATE POLICY "Users can create join requests"
ON room_join_requests FOR INSERT
TO public
WITH CHECK (user_id = auth.uid() OR EXISTS (
    SELECT 1 FROM users WHERE users.id = room_join_requests.user_id AND users.is_anonymous = TRUE
));

CREATE POLICY "Users can view their own requests"
ON room_join_requests FOR SELECT
TO public
USING (user_id = auth.uid() OR EXISTS (
    SELECT 1 FROM users WHERE users.id = room_join_requests.user_id AND users.is_anonymous = TRUE
));

CREATE POLICY "Room owners can view requests for their rooms"
ON room_join_requests FOR SELECT
TO public
USING (room_id IN (
    SELECT id FROM rooms WHERE owner_id = auth.uid()
));

CREATE POLICY "Room owners can update requests for their rooms"
ON room_join_requests FOR UPDATE
TO public
USING (room_id IN (
    SELECT id FROM rooms WHERE owner_id = auth.uid()
));

-- Create trigger for automatic updated_at
CREATE TRIGGER IF NOT EXISTS update_room_join_requests_updated_at
    BEFORE UPDATE ON room_join_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Migration V2 completed successfully';
    RAISE NOTICE 'Room join requests table created';
    RAISE NOTICE '===========================================';
END $$;

