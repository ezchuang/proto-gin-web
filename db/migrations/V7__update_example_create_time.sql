-- Switch covers to local /static assets for offline demos
UPDATE post 
SET created_at = NOW() - INTERVAL '10 days' - INTERVAL '14 hours', 
    updated_at = NOW() - INTERVAL '10 days' + INTERVAL '6 hours'
WHERE slug = 'welcome';

