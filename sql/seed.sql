-- Couple Card Game - Sample Data
-- This file provides initial categories and questions in multiple languages

-- Insert Categories
INSERT INTO categories (id, key, icon) VALUES
('00000000-0000-0000-0000-000000000001', 'relationship', '💕'),
('00000000-0000-0000-0000-000000000002', 'fun', '🎉'),
('00000000-0000-0000-0000-000000000003', 'deep', '🤔'),
('00000000-0000-0000-0000-000000000004', 'future', '🔮'),
('00000000-0000-0000-0000-000000000005', 'intimate', '❤️')
ON CONFLICT (key) DO NOTHING;

-- Insert Category Translations
INSERT INTO translations (language_code, translation_key, translation_value) VALUES
-- English
('en', 'category.relationship', 'Relationship'),
('en', 'category.fun', 'Fun & Games'),
('en', 'category.deep', 'Deep Thoughts'),
('en', 'category.future', 'Future Plans'),
('en', 'category.intimate', 'Intimate'),
-- French
('fr', 'category.relationship', 'Relation'),
('fr', 'category.fun', 'Amusement'),
('fr', 'category.deep', 'Réflexions Profondes'),
('fr', 'category.future', 'Plans Futurs'),
('fr', 'category.intimate', 'Intime'),
-- Spanish
('es', 'category.relationship', 'Relación'),
('es', 'category.fun', 'Diversión'),
('es', 'category.deep', 'Pensamientos Profundos'),
('es', 'category.future', 'Planes Futuros'),
('es', 'category.intimate', 'Íntimo')
ON CONFLICT (language_code, translation_key) DO NOTHING;

-- Insert Sample Questions (English)
INSERT INTO questions (category_id, language_code, text) VALUES
-- Relationship questions
('00000000-0000-0000-0000-000000000001', 'en', 'What was your first impression of me?'),
('00000000-0000-0000-0000-000000000001', 'en', 'What is your favorite memory of us together?'),
('00000000-0000-0000-0000-000000000001', 'en', 'What makes you feel most loved by me?'),
('00000000-0000-0000-0000-000000000001', 'en', 'How do you think we''ve grown as a couple?'),
('00000000-0000-0000-0000-000000000001', 'en', 'What is one thing you admire about our relationship?'),

-- Fun questions
('00000000-0000-0000-0000-000000000002', 'en', 'If you could have any superpower, what would it be?'),
('00000000-0000-0000-0000-000000000002', 'en', 'What would be your perfect weekend adventure?'),
('00000000-0000-0000-0000-000000000002', 'en', 'If we could go anywhere right now, where would you choose?'),
('00000000-0000-0000-0000-000000000002', 'en', 'What''s the funniest thing that''s happened to us?'),
('00000000-0000-0000-0000-000000000002', 'en', 'If you could learn any skill instantly, what would it be?'),

-- Deep questions
('00000000-0000-0000-0000-000000000003', 'en', 'What does happiness mean to you?'),
('00000000-0000-0000-0000-000000000003', 'en', 'What is one fear you''d like to overcome?'),
('00000000-0000-0000-0000-000000000003', 'en', 'How do you define success in life?'),
('00000000-0000-0000-0000-000000000003', 'en', 'What is something you''re grateful for today?'),
('00000000-0000-0000-0000-000000000003', 'en', 'What legacy do you want to leave behind?'),

-- Future questions
('00000000-0000-0000-0000-000000000004', 'en', 'Where do you see us in 5 years?'),
('00000000-0000-0000-0000-000000000004', 'en', 'What''s one goal you want us to achieve together?'),
('00000000-0000-0000-0000-000000000004', 'en', 'What kind of home would you love for us to have?'),
('00000000-0000-0000-0000-000000000004', 'en', 'What tradition do you want to start with our family?'),
('00000000-0000-0000-0000-000000000004', 'en', 'How do you imagine our life together changing?'),

-- Intimate questions
('00000000-0000-0000-0000-000000000005', 'en', 'What makes you feel closest to me?'),
('00000000-0000-0000-0000-000000000005', 'en', 'How can I better support you emotionally?'),
('00000000-0000-0000-0000-000000000005', 'en', 'What''s something you''ve never told me?'),
('00000000-0000-0000-0000-000000000005', 'en', 'What does intimacy mean to you?'),
('00000000-0000-0000-0000-000000000005', 'en', 'How can we deepen our connection?');

-- Success message
DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE 'Sample data inserted successfully!';
    RAISE NOTICE 'Database is ready to use';
    RAISE NOTICE '===========================================';
END $$;

