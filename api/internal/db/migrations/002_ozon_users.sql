-- Журнал отчётов Ozon: одна строка на день продаж. Отчёт за прошедший
-- день скачивается по расписанию, сегодняшний воркер докачивает в течение дня.
CREATE TABLE ozon_reports (
    day          date PRIMARY KEY,
    report_id    text NOT NULL,
    status       text NOT NULL CHECK (status IN ('queued', 'processing', 'success', 'error')),
    trigger      text NOT NULL CHECK (trigger IN ('schedule', 'manual')),
    attempts     integer NOT NULL DEFAULT 0,
    rows_total   integer NOT NULL DEFAULT 0, -- все заказы в отчёте
    orders       integer NOT NULL DEFAULT 0, -- из них с промокодами блогеров
    revenue_rub  bigint  NOT NULL DEFAULT 0,
    error        text,                       -- последняя ошибка, даже если повтор прошёл
    requested_at timestamptz NOT NULL DEFAULT now(),
    finished_at  timestamptz
);
CREATE INDEX ozon_reports_queue_idx ON ozon_reports (requested_at) WHERE status = 'queued';

CREATE TABLE users (
    id           bigserial PRIMARY KEY,
    name         text NOT NULL,
    email        text NOT NULL UNIQUE,
    role         text NOT NULL CHECK (role IN ('admin', 'marketer', 'manager', 'analyst', 'finance')),
    status       text NOT NULL CHECK (status IN ('active', 'invited', 'blocked')),
    last_seen_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
