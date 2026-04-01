-- Player location tracking for MMO-style shared world

CREATE TABLE IF NOT EXISTS player_locations (
    user_id         VARCHAR(128) PRIMARY KEY,
    character_id    VARCHAR(200) NOT NULL DEFAULT '',
    character_name  VARCHAR(200) NOT NULL,
    character_class VARCHAR(50) NOT NULL,
    character_level INT NOT NULL DEFAULT 1,
    region_x        INT NOT NULL DEFAULT 0,
    region_y        INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_locations_region ON player_locations(region_x, region_y);
CREATE INDEX IF NOT EXISTS idx_player_locations_updated ON player_locations(updated_at);
