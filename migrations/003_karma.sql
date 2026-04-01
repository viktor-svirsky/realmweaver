-- Add karma column to player_locations for the karma system
ALTER TABLE player_locations ADD COLUMN IF NOT EXISTS karma INT NOT NULL DEFAULT 0;
