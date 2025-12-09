-- Seed data for Couple Card Game
-- Run this after schema.sql

-- ============================================================================
-- ADMIN USER SETUP
-- ============================================================================
--
-- IMPORTANT: To create the admin user with email/password authentication:
--
-- Option 1: Use Supabase Dashboard (Production/Hosted)
--   1. Go to Supabase Dashboard → Authentication → Users
--   2. Click "Add User" → "Create new user"
--   3. Enter:
--      - Email: admin@example.com
--      - Password: admin123 (change in production!)
--      - Set user_metadata: {"username": "admin"}
--   4. Copy the user UUID
--   5. Update the INSERT below with that UUID
--
-- Option 2: Use Signup Form (Recommended for local development)
--   1. Start the server: make dev
--   2. Navigate to: http://localhost:8080/auth/signup
--   3. Create account with:
--      - Username: admin
--      - Email: admin@example.com
--      - Password: admin123
--   4. This automatically creates the user in both Supabase Auth and application DB
--   5. Update the user to admin: UPDATE users SET is_admin = TRUE WHERE email = 'admin@example.com';
--
-- Option 3: Dev Login (Development Only)
--   - Visit /auth/dev-login-admin to login as the seeded admin user
--   - This bypasses authentication (development only!)
--
-- ============================================================================

-- Insert admin user profile (authentication credentials stored in Supabase Auth)
INSERT INTO users (id, email, username, is_admin, is_anonymous) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', 'admin', TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

-- Insert categories (with human-readable labels)
INSERT INTO categories (id, key, label) VALUES
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'couples', 'Couples'),
('b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'friends', 'Friends'),
('b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'sex', 'Intimacy'),
('b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'family', 'Family'),
('b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'deep', 'Deep Questions'),
('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fun', 'Fun')
ON CONFLICT (key) DO NOTHING;

-- Insert sample questions in English (base_question_id = id for English questions)
INSERT INTO questions (id, category_id, lang_code, question_text, base_question_id) VALUES
-- Couples questions
('d1a2b3c4-5e6f-47a8-89b0-1c2d3e4f5a6b', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What was your first impression of me?', 'd1a2b3c4-5e6f-47a8-89b0-1c2d3e4f5a6b'),
('e2b3c4d5-6f7a-48b9-90c1-2d3e4f5a6b7c', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite memory of us together?', 'e2b3c4d5-6f7a-48b9-90c1-2d3e4f5a6b7c'),
('f3c4d5e6-7a8b-49ca-91d2-3e4f5a6b7c8d', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you love most about our relationship?', 'f3c4d5e6-7a8b-49ca-91d2-3e4f5a6b7c8d'),
('a4d5e6f7-8b9c-4adb-92e3-4f5a6b7c8d9e', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something I do that makes you feel loved?', 'a4d5e6f7-8b9c-4adb-92e3-4f5a6b7c8d9e'),
('b5e6f7a8-9cad-4bec-93f4-5a6b7c8d9eaf', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'Where do you see us in 5 years?', 'b5e6f7a8-9cad-4bec-93f4-5a6b7c8d9eaf'),
('c6f7a8b9-adbe-4cfd-94a5-6b7c8d9eafba', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is one thing you would change about our relationship?', 'c6f7a8b9-adbe-4cfd-94a5-6b7c8d9eafba'),
('d7a8b9ca-becf-4dae-95b6-7c8d9eafbacb', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite way to spend time together?', 'd7a8b9ca-becf-4dae-95b6-7c8d9eafbacb'),
('e8b9cadb-cfda-4ebf-96c7-8d9eafbacbdc', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most connected to me?', 'e8b9cadb-cfda-4ebf-96c7-8d9eafbacbdc'),

-- Friends questions
('f9cadebc-daeb-4fca-97d8-9eafbacbdced', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite thing about our friendship?', 'f9cadebc-daeb-4fca-97d8-9eafbacbdced'),
('aabefccd-ebfc-4adb-98e9-afbacbdcedfe', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the funniest memory we share?', 'aabefccd-ebfc-4adb-98e9-afbacbdcedfe'),
('bbcfadde-fcad-4bec-99fa-bacbdcedefaf', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could give me one piece of advice, what would it be?', 'bbcfadde-fcad-4bec-99fa-bacbdcedefaf'),
('ccdabeef-adbe-4cfd-9aab-cbdcedefafba', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think I am really good at?', 'ccdabeef-adbe-4cfd-9aab-cbdcedefafba'),
('ddebcffa-becf-4dae-9bbc-dcedefafbacb', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something we should do together more often?', 'ddebcffa-becf-4dae-9bbc-dcedefafbacb'),

-- Intimate questions
('eefcdaab-cfda-4ebf-9ccd-edefafbacbdc', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something new you would like to try together?', 'eefcdaab-cfda-4ebf-9ccd-edefafbacbdc'),
('ffadebbc-daeb-4fca-9dde-efafbacbdced', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most attractive?', 'ffadebbc-daeb-4fca-9dde-efafbacbdced'),
('aabefccd-ebfc-4adb-9eef-fabacbdcedfe', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest turn on?', 'aabefccd-ebfc-4adb-9eef-fabacbdcedfe'),
('bbcfadde-fcad-4bec-9ffa-abacbdcedefb', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is a fantasy you have never shared?', 'bbcfadde-fcad-4bec-9ffa-abacbdcedefb'),

-- Family questions
('ccdabeef-adbe-4cfd-9aab-bacbdcedefac', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite family tradition?', 'ccdabeef-adbe-4cfd-9aab-bacbdcedefac'),
('ddebcffa-becf-4dae-9bbc-cbdcedefafbd', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What values do you want to pass on to your children?', 'ddebcffa-becf-4dae-9bbc-cbdcedefafbd'),
('eefcdaab-cfda-4ebf-9ccd-dcedefafbace', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something you learned from your parents?', 'eefcdaab-cfda-4ebf-9ccd-dcedefafbace'),
('ffadebbc-daeb-4fca-9dde-edefafbacbdf', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'How do you want to celebrate special occasions as a family?', 'ffadebbc-daeb-4fca-9dde-edefafbacbdf'),

-- Deep questions
('11befccd-ebfc-4adb-9eef-efafbacbdcea', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest fear?', '11befccd-ebfc-4adb-9eef-efafbacbdcea'),
('22cfadde-fcad-4bec-9ffa-fabacbdcedfb', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the most important lesson life has taught you?', '22cfadde-fcad-4bec-9ffa-fabacbdcedfb'),
('33dabeef-adbe-4cfd-9aab-abacbdcedefc', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think happens after we die?', '33dabeef-adbe-4cfd-9aab-abacbdcedefc'),
('44ebcffa-becf-4dae-9bbc-bacbdcedefad', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could change one thing about your past, what would it be?', '44ebcffa-becf-4dae-9bbc-bacbdcedefad'),
('55fcdaab-cfda-4ebf-9ccd-cbdcedefafbe', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What gives your life meaning?', '55fcdaab-cfda-4ebf-9ccd-cbdcedefafbe'),

-- Fun questions
('66adebbc-daeb-4fca-9dde-dcedefafbacf', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could have any superpower, what would it be?', '66adebbc-daeb-4fca-9dde-dcedefafbacf'),
('77befccd-ebfc-4adb-9eef-edefafbacbda', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your guilty pleasure?', '77befccd-ebfc-4adb-9eef-edefafbacbda'),
('88cfadde-fcad-4bec-9ffa-efafbacbdceb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you won the lottery, what is the first thing you would do?', '88cfadde-fcad-4bec-9ffa-efafbacbdceb'),
('99dabeef-adbe-4cfd-9aab-fabacbdcedef', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite movie and why?', '99dabeef-adbe-4cfd-9aab-fabacbdcedef'),
('aaebcffa-becf-4dae-9bbc-abacbdcedefb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could travel anywhere in the world, where would you go?', 'aaebcffa-becf-4dae-9bbc-abacbdcedefb')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in French (base_question_id = English question ID)
INSERT INTO questions (id, category_id, lang_code, question_text, base_question_id) VALUES
-- Couples questions
('bbfcdaab-cfda-4ebf-9ccd-bacbdcedefac', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle a été votre première impression de moi?', 'd1a2b3c4-5e6f-47a8-89b0-1c2d3e4f5a6b'),
('ccadebbc-daeb-4fca-9dde-cbdcedefafbd', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre souvenir préféré de nous ensemble?', 'e2b3c4d5-6f7a-48b9-90c1-2d3e4f5a6b7c'),
('ddbefccd-ebfc-4adb-9eef-dcedefafbace', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Qu''aimez-vous le plus dans notre relation?', 'f3c4d5e6-7a8b-49ca-91d2-3e4f5a6b7c8d'),

-- Friends questions
('eecfadde-fcad-4bec-9ffa-edefafbacbdf', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle est votre chose préférée dans notre amitié?', 'f9cadebc-daeb-4fca-97d8-9eafbacbdced'),
('ffdabeef-adbe-4cfd-9aab-efafbacbdcea', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est le souvenir le plus drôle que nous partageons?', 'aabefccd-ebfc-4adb-98e9-afbacbdcedfe'),

-- Fun questions
('11ebcffa-becf-4dae-9bbc-fabacbdcedfb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Si vous pouviez avoir un super pouvoir, lequel serait-ce?', '66adebbc-daeb-4fca-9dde-dcedefafbacf'),
('22fcdaab-cfda-4ebf-9ccd-abacbdcedefc', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre plaisir coupable?', '77befccd-ebfc-4adb-9eef-edefafbacbda')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in Japanese (base_question_id = English question ID)
INSERT INTO questions (id, category_id, lang_code, question_text, base_question_id) VALUES
-- Couples questions
('33adebbc-daeb-4fca-9dde-bacbdcedefad', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私の第一印象は何でしたか？', 'd1a2b3c4-5e6f-47a8-89b0-1c2d3e4f5a6b'),
('44befccd-ebfc-4adb-9eef-cbdcedefafbe', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '一緒に過ごしたお気に入りの思い出は何ですか？', 'e2b3c4d5-6f7a-48b9-90c1-2d3e4f5a6b7c'),
('55cfadde-fcad-4bec-9ffa-dcedefafbacf', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私たちの関係で一番好きなことは何ですか？', 'f3c4d5e6-7a8b-49ca-91d2-3e4f5a6b7c8d'),

-- Fun questions
('66dabeef-adbe-4cfd-9aab-edefafbacbda', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'もし超能力を持てるなら、どんな力が欲しいですか？', '66adebbc-daeb-4fca-9dde-dcedefafbacf'),
('77ebcffa-becf-4dae-9bbc-efafbacbdceb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'あなたの罪悪感のある楽しみは何ですか？', '77befccd-ebfc-4adb-9eef-edefafbacbda')
ON CONFLICT (id) DO NOTHING;

-- Insert base UI translations
INSERT INTO translations (lang_code, key, value) VALUES
('en', 'nav.home', 'Home'),
('en', 'nav.login', 'Login'),
('en', 'nav.logout', 'Logout'),
('fr', 'nav.home', 'Accueil'),
('fr', 'nav.login', 'Connexion'),
('fr', 'nav.logout', 'Déconnexion'),
('ja', 'nav.home', 'ホーム'),
('ja', 'nav.login', 'ログイン'),
('ja', 'nav.logout', 'ログアウト')
ON CONFLICT (lang_code, key) DO NOTHING;

