CREATE TABLE IF NOT EXISTS subscription_plans (
    id                      VARCHAR(20) PRIMARY KEY,
    name                    VARCHAR(50) NOT NULL,
    price_kop               INTEGER NOT NULL DEFAULT 0,
    period_days             INTEGER NOT NULL DEFAULT 30,
    max_prompts             INTEGER NOT NULL DEFAULT 50,
    max_collections         INTEGER NOT NULL DEFAULT 3,
    max_ai_requests_daily   INTEGER NOT NULL DEFAULT 5,
    ai_requests_is_total    BOOLEAN NOT NULL DEFAULT FALSE,
    max_teams               INTEGER NOT NULL DEFAULT 1,
    max_team_members        INTEGER NOT NULL DEFAULT 3,
    max_share_links         INTEGER NOT NULL DEFAULT 2,
    max_ext_uses_daily      INTEGER NOT NULL DEFAULT 5,
    max_mcp_uses_daily      INTEGER NOT NULL DEFAULT 5,
    features                JSONB NOT NULL DEFAULT '[]',
    sort_order              INTEGER NOT NULL DEFAULT 0,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO subscription_plans (id, name, price_kop, period_days,
    max_prompts, max_collections, max_ai_requests_daily, ai_requests_is_total,
    max_teams, max_team_members, max_share_links, max_ext_uses_daily,
    max_mcp_uses_daily, features, sort_order)
VALUES
    ('free', 'Free', 0, 0,
     50, 3, 5, TRUE,
     1, 3, 2, 5, 5,
     '[]'::jsonb, 0),
    ('pro', 'Pro', 59900, 30,
     500, -1, 10, FALSE,
     5, 10, 10, 30, 30,
     '["priority_support"]'::jsonb, 1),
    ('max', 'Max', 129900, 30,
     -1, -1, 30, FALSE,
     -1, -1, -1, -1, -1,
     '["priority_support"]'::jsonb, 2);
