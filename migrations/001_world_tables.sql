-- Realmweaver world data schema

CREATE TABLE IF NOT EXISTS regions (
    id          SERIAL PRIMARY KEY,
    x           INT NOT NULL,
    y           INT NOT NULL,
    biome       VARCHAR(50) NOT NULL,
    difficulty  INT NOT NULL DEFAULT 1,
    name        VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    lore        TEXT NOT NULL DEFAULT '',
    structures  JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(x, y)
);

CREATE INDEX IF NOT EXISTS idx_regions_coords ON regions(x, y);

CREATE TABLE IF NOT EXISTS factions (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    region_ids  JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS npcs (
    id                SERIAL PRIMARY KEY,
    region_id         INT NOT NULL REFERENCES regions(id),
    name              VARCHAR(200) NOT NULL,
    race              VARCHAR(50) NOT NULL DEFAULT 'human',
    occupation        VARCHAR(100) NOT NULL DEFAULT '',
    personality_prompt TEXT NOT NULL DEFAULT '',
    disposition       JSONB NOT NULL DEFAULT '{}',
    memory_tags       JSONB NOT NULL DEFAULT '[]',
    faction_id        INT REFERENCES factions(id),
    alive             BOOLEAN NOT NULL DEFAULT TRUE,
    location_detail   VARCHAR(200) NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_npcs_region ON npcs(region_id);

CREATE TABLE IF NOT EXISTS faction_reputation (
    faction_id  INT NOT NULL REFERENCES factions(id),
    user_id     VARCHAR(128) NOT NULL,
    score       INT NOT NULL DEFAULT 0,
    PRIMARY KEY (faction_id, user_id)
);

CREATE TABLE IF NOT EXISTS quests (
    id            SERIAL PRIMARY KEY,
    title         VARCHAR(300) NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    region_id     INT NOT NULL REFERENCES regions(id),
    giver_npc_id  INT REFERENCES npcs(id),
    objectives    JSONB NOT NULL DEFAULT '[]',
    rewards       JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quests_region ON quests(region_id);

CREATE TABLE IF NOT EXISTS quest_progress (
    quest_id    INT NOT NULL REFERENCES quests(id),
    user_id     VARCHAR(128) NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT 'active',
    progress    JSONB NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (quest_id, user_id)
);

CREATE TABLE IF NOT EXISTS world_events (
    id                  SERIAL PRIMARY KEY,
    event_type          VARCHAR(100) NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    affected_region_ids JSONB NOT NULL DEFAULT '[]',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved            BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS player_messages (
    id          SERIAL PRIMARY KEY,
    region_id   INT NOT NULL REFERENCES regions(id),
    user_id     VARCHAR(128) NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_messages_region ON player_messages(region_id);

CREATE TABLE IF NOT EXISTS combat_log (
    id          SERIAL PRIMARY KEY,
    user_id     VARCHAR(128) NOT NULL,
    region_id   INT NOT NULL REFERENCES regions(id),
    entries     JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_combat_log_user ON combat_log(user_id);
