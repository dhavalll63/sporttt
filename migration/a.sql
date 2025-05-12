INSERT INTO roles (name, created_at, updated_at)
VALUES 
  ('admin', NOW(), NOW()),
  ('player', NOW(), NOW()),
  ('venue_manager', NOW(), NOW()),
  ('match_creator', NOW(), NOW()),
  ('scorer', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
