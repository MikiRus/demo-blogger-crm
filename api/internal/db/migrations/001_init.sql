CREATE TABLE platforms (
    id   smallserial PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL
);

CREATE TABLE bloggers (
    id          bigserial PRIMARY KEY,
    name        text     NOT NULL,
    handle      text     NOT NULL UNIQUE,
    platform_id smallint NOT NULL REFERENCES platforms (id),
    followers   integer  NOT NULL CHECK (followers >= 0),
    avg_reach   integer  NOT NULL CHECK (avg_reach >= 0),
    category    text     NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Рекламная кампания: продвижение одного товара через пул блогеров.
CREATE TABLE campaigns (
    id         bigserial PRIMARY KEY,
    title      text   NOT NULL,
    product    text   NOT NULL,
    status     text   NOT NULL CHECK (status IN ('planned', 'active', 'finished')),
    start_date date   NOT NULL,
    end_date   date   NOT NULL CHECK (end_date >= start_date),
    budget_rub bigint NOT NULL CHECK (budget_rub >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Интеграция: конкретный пост блогера в рамках кампании.
CREATE TABLE integrations (
    id           bigserial PRIMARY KEY,
    campaign_id  bigint NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    blogger_id   bigint NOT NULL REFERENCES bloggers (id) ON DELETE CASCADE,
    publish_date date   NOT NULL,
    cost_rub     bigint NOT NULL CHECK (cost_rub >= 0),
    promo_code   text   NOT NULL UNIQUE,
    utm_campaign text   NOT NULL,
    status       text   NOT NULL CHECK (status IN ('planned', 'published'))
);
CREATE INDEX integrations_campaign_idx ON integrations (campaign_id);
CREATE INDEX integrations_blogger_idx ON integrations (blogger_id);

-- Дневные продажи по интеграции в разрезе источника.
-- ozon и site (CRM) подтягивает воркер.
CREATE TABLE integration_metrics (
    integration_id bigint NOT NULL REFERENCES integrations (id) ON DELETE CASCADE,
    day            date   NOT NULL,
    source         text   NOT NULL CHECK (source IN ('ozon', 'wb', 'site')),
    orders         integer NOT NULL CHECK (orders >= 0),
    revenue_rub    bigint  NOT NULL CHECK (revenue_rub >= 0),
    PRIMARY KEY (integration_id, day, source)
);
CREATE INDEX integration_metrics_day_idx ON integration_metrics (day);
