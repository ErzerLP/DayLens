-- DayLens 数据库初始化
-- 用于 PostgreSQL 16

-- 活动记录表
CREATE TABLE IF NOT EXISTS activities (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             INT NOT NULL DEFAULT 1,
    client_id           VARCHAR(64) NOT NULL DEFAULT '',
    client_ts           BIGINT NOT NULL,
    timestamp           BIGINT NOT NULL,
    app_name            VARCHAR(255) NOT NULL,
    window_title        TEXT NOT NULL,
    screenshot_key      VARCHAR(512) DEFAULT '',
    ocr_text            TEXT,
    category            VARCHAR(64) NOT NULL,
    semantic_category   VARCHAR(64),
    semantic_confidence INT,
    duration            INT NOT NULL,
    browser_url         TEXT,
    executable_path     TEXT,
    extra_json          JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_activity_idempotent UNIQUE (user_id, client_id, client_ts)
);

CREATE INDEX IF NOT EXISTS idx_activities_timestamp ON activities (timestamp);
CREATE INDEX IF NOT EXISTS idx_activities_user_date ON activities (user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_activities_app ON activities (app_name);
CREATE INDEX IF NOT EXISTS idx_activities_category ON activities (category);

-- 日报表
CREATE TABLE IF NOT EXISTS daily_reports (
    id          BIGSERIAL PRIMARY KEY,
    user_id     INT NOT NULL DEFAULT 1,
    date        DATE NOT NULL,
    content     TEXT NOT NULL,
    ai_mode     VARCHAR(32) NOT NULL,
    model_name  VARCHAR(128),
    used_ai     BOOLEAN NOT NULL DEFAULT FALSE,
    extra_json  JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_daily_report_date UNIQUE (user_id, date)
);

-- 小时摘要表
CREATE TABLE IF NOT EXISTS hourly_summaries (
    id                          BIGSERIAL PRIMARY KEY,
    user_id                     INT NOT NULL DEFAULT 1,
    date                        DATE NOT NULL,
    hour                        INT NOT NULL CHECK (hour >= 0 AND hour <= 23),
    summary                     TEXT NOT NULL,
    main_apps                   TEXT NOT NULL,
    activity_count              INT NOT NULL,
    total_duration              INT NOT NULL,
    representative_screenshots  TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_hourly UNIQUE (user_id, date, hour)
);

-- 分类规则表
CREATE TABLE IF NOT EXISTS category_rules (
    id          BIGSERIAL PRIMARY KEY,
    user_id     INT NOT NULL DEFAULT 1,
    app_name    VARCHAR(255) NOT NULL,
    category    VARCHAR(64) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_category_rule UNIQUE (user_id, app_name)
);
