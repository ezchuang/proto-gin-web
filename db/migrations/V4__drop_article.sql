-- Remove legacy article artifacts

DROP TABLE IF EXISTS article;

DELETE FROM app_user WHERE email = 'demo@example.com';
