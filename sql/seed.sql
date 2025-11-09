-- Seed data for Couple Card Game
-- Run this after schema.sql

-- Insert admin user (password: admin123 - change this in production!)
INSERT INTO users (id, email, name, username, is_admin, is_anonymous) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', 'Admin', 'admin', TRUE, FALSE)
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

-- Insert sample questions in English
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('d1a2b3c4-5e6f-47a8-89b0-1c2d3e4f5a6b', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What was your first impression of me?'),
('e2b3c4d5-6f7a-48b9-90c1-2d3e4f5a6b7c', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite memory of us together?'),
('f3c4d5e6-7a8b-49ca-91d2-3e4f5a6b7c8d', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you love most about our relationship?'),
('a4d5e6f7-8b9c-4adb-92e3-4f5a6b7c8d9e', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something I do that makes you feel loved?'),
('b5e6f7a8-9cad-4bec-93f4-5a6b7c8d9eaf', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'Where do you see us in 5 years?'),
('c6f7a8b9-adbe-4cfd-94a5-6b7c8d9eafba', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is one thing you would change about our relationship?'),
('d7a8b9ca-becf-4dae-95b6-7c8d9eafbacb', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite way to spend time together?'),
('e8b9cadb-cfda-4ebf-96c7-8d9eafbacbdc', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most connected to me?'),

-- Friends questions
('f9cadebc-daeb-4fca-97d8-9eafbacbdced', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite thing about our friendship?'),
('aabefccd-ebfc-4adb-98e9-afbacbdcedfe', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the funniest memory we share?'),
('bbcfadde-fcad-4bec-99fa-bacbdcedefaf', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could give me one piece of advice, what would it be?'),
('ccdabeef-adbe-4cfd-9aab-cbdcedefafba', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think I am really good at?'),
('ddebcffa-becf-4dae-9bbc-dcedefafbacb', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something we should do together more often?'),

-- Intimate questions
('eefcdaab-cfda-4ebf-9ccd-edefafbacbdc', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something new you would like to try together?'),
('ffadebbc-daeb-4fca-9dde-efafbacbdced', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What makes you feel most attractive?'),
('aabefccd-ebfc-4adb-9eef-fabacbdcedfe', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest turn on?'),
('bbcfadde-fcad-4bec-9ffa-abacbdcedefb', 'b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is a fantasy you have never shared?'),

-- Family questions
('ccdabeef-adbe-4cfd-9aab-bacbdcedefac', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite family tradition?'),
('ddebcffa-becf-4dae-9bbc-cbdcedefafbd', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What values do you want to pass on to your children?'),
('eefcdaab-cfda-4ebf-9ccd-dcedefafbace', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is something you learned from your parents?'),
('ffadebbc-daeb-4fca-9dde-edefafbacbdf', 'b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'How do you want to celebrate special occasions as a family?'),

-- Deep questions
('11befccd-ebfc-4adb-9eef-efafbacbdcea', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your biggest fear?'),
('22cfadde-fcad-4bec-9ffa-fabacbdcedfb', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is the most important lesson life has taught you?'),
('33dabeef-adbe-4cfd-9aab-abacbdcedefc', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What do you think happens after we die?'),
('44ebcffa-becf-4dae-9bbc-bacbdcedefad', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could change one thing about your past, what would it be?'),
('55fcdaab-cfda-4ebf-9ccd-cbdcedefafbe', 'b5eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What gives your life meaning?'),

-- Fun questions
('66adebbc-daeb-4fca-9dde-dcedefafbacf', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could have any superpower, what would it be?'),
('77befccd-ebfc-4adb-9eef-edefafbacbda', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your guilty pleasure?'),
('88cfadde-fcad-4bec-9ffa-efafbacbdceb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you won the lottery, what is the first thing you would do?'),
('99dabeef-adbe-4cfd-9aab-fabacbdcedef', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'What is your favorite movie and why?'),
('aaebcffa-becf-4dae-9bbc-abacbdcedefb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'en', 'If you could travel anywhere in the world, where would you go?')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in French
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('bbfcdaab-cfda-4ebf-9ccd-bacbdcedefac', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle a été votre première impression de moi?'),
('ccadebbc-daeb-4fca-9dde-cbdcedefafbd', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre souvenir préféré de nous ensemble?'),
('ddbefccd-ebfc-4adb-9eef-dcedefafbace', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Qu''aimez-vous le plus dans notre relation?'),

-- Friends questions
('eecfadde-fcad-4bec-9ffa-edefafbacbdf', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quelle est votre chose préférée dans notre amitié?'),
('ffdabeef-adbe-4cfd-9aab-efafbacbdcea', 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est le souvenir le plus drôle que nous partageons?'),

-- Fun questions
('11ebcffa-becf-4dae-9bbc-fabacbdcedfb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Si vous pouviez avoir un super pouvoir, lequel serait-ce?'),
('22fcdaab-cfda-4ebf-9ccd-abacbdcedefc', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'fr', 'Quel est votre plaisir coupable?')
ON CONFLICT (id) DO NOTHING;

-- Insert sample questions in Japanese
INSERT INTO questions (id, category_id, lang_code, question_text) VALUES
-- Couples questions
('33adebbc-daeb-4fca-9dde-bacbdcedefad', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私の第一印象は何でしたか？'),
('44befccd-ebfc-4adb-9eef-cbdcedefafbe', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '一緒に過ごしたお気に入りの思い出は何ですか？'),
('55cfadde-fcad-4bec-9ffa-dcedefafbacf', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', '私たちの関係で一番好きなことは何ですか？'),

-- Fun questions
('66dabeef-adbe-4cfd-9aab-edefafbacbda', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'もし超能力を持てるなら、どんな力が欲しいですか？'),
('77ebcffa-becf-4dae-9bbc-efafbacbdceb', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ja', 'あなたの罪悪感のある楽しみは何ですか？')
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

