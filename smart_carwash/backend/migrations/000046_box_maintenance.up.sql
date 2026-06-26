-- Учёт моточасов боксов-моек для регламентного ТО (каждые 500 моточасов).
-- Считаем проданное время сессий на боксе, конвертируем в моточасы.
-- baseline_minutes = проданные минуты на момент последнего ТО (точка отсчёта).

CREATE TABLE IF NOT EXISTS box_maintenance (
    box_number       INTEGER PRIMARY KEY,
    last_to_at       TIMESTAMPTZ NOT NULL,
    baseline_minutes BIGINT NOT NULL DEFAULT 0,
    alerted          BOOLEAN NOT NULL DEFAULT false,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Инициализация для моечных боксов (1-8, 11-18). Базовое ТО — 20 мая 2026.
INSERT INTO box_maintenance (box_number, last_to_at, baseline_minutes, alerted)
VALUES
  (1,'2026-05-20 00:00:00+00',0,false),(2,'2026-05-20 00:00:00+00',0,false),
  (3,'2026-05-20 00:00:00+00',0,false),(4,'2026-05-20 00:00:00+00',0,false),
  (5,'2026-05-20 00:00:00+00',0,false),(6,'2026-05-20 00:00:00+00',0,false),
  (7,'2026-05-20 00:00:00+00',0,false),(8,'2026-05-20 00:00:00+00',0,false),
  (11,'2026-05-20 00:00:00+00',0,false),(12,'2026-05-20 00:00:00+00',0,false),
  (13,'2026-05-20 00:00:00+00',0,false),(14,'2026-05-20 00:00:00+00',0,false),
  (15,'2026-05-20 00:00:00+00',0,false),(16,'2026-05-20 00:00:00+00',0,false),
  (17,'2026-05-20 00:00:00+00',0,false),(18,'2026-05-20 00:00:00+00',0,false)
ON CONFLICT (box_number) DO NOTHING;

-- baseline = проданные минуты бокса ДО даты базового ТО (20.05.2026)
UPDATE box_maintenance bm
SET baseline_minutes = COALESCE((
    SELECT SUM(s.rental_time_minutes + s.extension_time_minutes)
    FROM sessions s
    JOIN wash_boxes w ON w.id = s.box_id
    WHERE w.number = bm.box_number
      AND s.status = 'complete'
      AND s.created_at < bm.last_to_at
), 0);
