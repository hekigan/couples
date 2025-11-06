-- Seed data for Couple Card Game
-- Run this after schema.sql

-- Insert admin user (password: admin123 - change this in production!)
INSERT INTO users (id, email, name, is_admin, is_anonymous) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', 'Admin', TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

-- Insert categories
INSERT INTO categories (id, key) VALUES
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'couples'),
('b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'friends'),
('b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'sex'),
('b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'family'),
('b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'deep'),
('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fun')
ON CONFLICT (key) DO NOTHING;

-- Insert sample questions in English
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('c1000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What was your first impression of me?'),
('c1000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite memory of us together?'),
('c1000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you love most about our relationship?'),
('c1000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something I do that makes you feel loved?'),
('c1000005-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'Where do you see us in 5 years?'),
('c1000006-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is one thing you would change about our relationship?'),
('c1000007-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite way to spend time together?'),
('c1000008-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most connected to me?'),

-- Friends questions
('c2000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite thing about our friendship?'),
('c2000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the funniest memory we share?'),
('c2000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could give me one piece of advice, what would it be?'),
('c2000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think I am really good at?'),
('c2000005-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something we should do together more often?'),

-- Intimate questions
('c3000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something new you would like to try together?'),
('c3000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most attractive?'),
('c3000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest turn on?'),
('c3000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is a fantasy you have never shared?'),

-- Family questions
('c4000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite family tradition?'),
('c4000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What values do you want to pass on to your children?'),
('c4000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something you learned from your parents?'),
('c4000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'How do you want to celebrate special occasions as a family?'),

-- Deep questions
('c5000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest fear?'),
('c5000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the most important lesson life has taught you?'),
('c5000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think happens after we die?'),
('c5000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could change one thing about your past, what would it be?'),
('c5000005-9c0b-4ef8-bb6d-6bb9bd380a11', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What gives your life meaning?'),

-- Fun questions
('c6000001-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could have any superpower, what would it be?'),
('c6000002-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your guilty pleasure?'),
('c6000003-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you won the lottery, what is the first thing you would do?'),
('c6000004-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite movie and why?'),
('c6000005-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could travel anywhere in the world, where would you go?')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in French
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('c1000101-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle a été votre première impression de moi?'),
('c1000102-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre souvenir préféré de nous ensemble?'),
('c1000103-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Qu''aimez-vous le plus dans notre relation?'),

-- Friends questions
('c2000101-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle est votre chose préférée dans notre amitié?'),
('c2000102-9c0b-4ef8-bb6d-6bb9bd380a11', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est le souvenir le plus drôle que nous partageons?'),

-- Fun questions
('c6000101-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Si vous pouviez avoir un super pouvoir, lequel serait-ce?'),
('c6000102-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre plaisir coupable?')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in Japanese
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('c1000201-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私の第一印象は何でしたか？'),
('c1000202-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '一緒に過ごしたお気に入りの思い出は何ですか？'),
('c1000203-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私たちの関係で一番好きなことは何ですか？'),

-- Fun questions
('c6000201-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'もし超能力を持てるなら、どんな力が欲しいですか？'),
('c6000202-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'あなたの罪悪感のある楽しみは何ですか？')
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


