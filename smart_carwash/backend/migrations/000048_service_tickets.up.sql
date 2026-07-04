-- Сервисные наряды (Фаза 1): справочник симптомов кассира, справочник работ мастера,
-- наряды на сервис (двухэтапные) и получатели push-уведомлений.
-- Спецификация: docs/maintenance_design.md, разделы 4, 8, 12-15.

-- Тип бокса: wash (1-8,11-18) | vacuum (9,10,19,20) | air (21,22,23).

-- ============ Справочник симптомов кассира ============
CREATE TABLE IF NOT EXISTS symptom_types (
    id          SERIAL PRIMARY KEY,
    group_name  VARCHAR(100) NOT NULL,           -- Вода/давление, Химия, Электрика, ...
    name        VARCHAR(255) NOT NULL,
    box_type    VARCHAR(20)  NOT NULL,           -- wash | vacuum | air
    sort_order  INTEGER      NOT NULL DEFAULT 0,
    is_active   BOOLEAN      NOT NULL DEFAULT true
);

-- ============ Справочник работ мастера (группа -> деталь) ============
CREATE TABLE IF NOT EXISTS component_groups (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    carrier     VARCHAR(10)  NOT NULL,           -- box | machine | any
    box_type    VARCHAR(20)  NOT NULL DEFAULT 'any', -- wash | vacuum | air | any
    sort_order  INTEGER      NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS component_types (
    id          SERIAL PRIMARY KEY,
    group_id    INTEGER      NOT NULL REFERENCES component_groups(id),
    name        VARCHAR(255) NOT NULL,
    sort_order  INTEGER      NOT NULL DEFAULT 0
);

-- ============ Наряды на сервис ============
CREATE TABLE IF NOT EXISTS service_tickets (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    box_id               UUID,
    box_number           INTEGER NOT NULL,
    box_type             VARCHAR(20),
    status               VARCHAR(20) NOT NULL DEFAULT 'open',  -- open | closed
    opened_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    opened_by            VARCHAR(255),        -- имя кассира
    opened_by_cashier_id UUID,                -- ссылка на кассира (если известна)
    symptom_id           INTEGER REFERENCES symptom_types(id),
    cashier_comment      TEXT,
    closed_at            TIMESTAMPTZ,
    closed_by            VARCHAR(255),         -- имя мастера
    master_comment       TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_tickets_status_box ON service_tickets (status, box_number);
CREATE INDEX IF NOT EXISTS idx_service_tickets_opened_at ON service_tickets (opened_at DESC);

-- ============ Выполненные работы в рамках наряда ============
CREATE TABLE IF NOT EXISTS ticket_works (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id    UUID NOT NULL REFERENCES service_tickets(id) ON DELETE CASCADE,
    component_id INTEGER NOT NULL REFERENCES component_types(id),
    action       VARCHAR(20) NOT NULL,   -- replace | repair
    carrier      VARCHAR(10),            -- box | machine | any (денормализовано для аналитики)
    motor_hours  INTEGER,                -- моточасы носителя на момент закрытия
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_works_ticket ON ticket_works (ticket_id);
CREATE INDEX IF NOT EXISTS idx_ticket_works_component ON ticket_works (component_id);

-- ============ Получатели push-уведомлений (управление в админке) ============
CREATE TABLE IF NOT EXISTS notification_recipients (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    chat_id     BIGINT NOT NULL,
    event_type  VARCHAR(50) NOT NULL DEFAULT 'service_ticket',
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ==================================================================
-- СИДЫ
-- ==================================================================

-- Симптомы кассира (раздел 12)
INSERT INTO symptom_types (group_name, name, box_type, sort_order) VALUES
    -- Мойка: вода/давление
    ('Вода / давление', 'Нет воды / давления (аппарат не запускается / слабый напор)', 'wash', 1),
    ('Вода / давление', 'Течь / разрыв шланга АВД', 'wash', 2),
    ('Вода / давление', 'Течёт пистолет / соединения', 'wash', 3),
    ('Вода / давление', 'Течь в аппаратной (шланг / труба)', 'wash', 4),
    ('Вода / давление', 'Аппарат включается и сразу отключается', 'wash', 5),
    -- Мойка: химия
    ('Химия', 'Химия не подаётся', 'wash', 6),
    ('Химия', 'Течь шланга химии', 'wash', 7),
    -- Мойка: электрика
    ('Электрика', 'Не горит свет в боксе', 'wash', 8),
    -- Пылесос
    ('Пылесос', 'Пылесос не включается', 'vacuum', 1),
    ('Пылесос', 'Пылесос не выключается', 'vacuum', 2),
    ('Пылесос', 'Слабое всасывание', 'vacuum', 3),
    ('Пылесос', 'Не горит свет в боксе', 'vacuum', 4),
    -- Воздух
    ('Воздух', 'Воздух не включается (электроклапан)', 'air', 1),
    ('Воздух', 'Воздух не выключается (электроклапан)', 'air', 2)
ON CONFLICT DO NOTHING;

-- Группы работ мастера (раздел 13)
INSERT INTO component_groups (name, carrier, box_type, sort_order) VALUES
    ('Пистолет моечный', 'box', 'wash', 1),
    ('Пистолет химии', 'box', 'wash', 2),
    ('Бокс', 'box', 'wash', 3),
    ('Аппаратная', 'box', 'wash', 4),
    ('Воздух', 'box', 'air', 5),
    ('Пылесос', 'box', 'vacuum', 6),
    ('Регулятор давления', 'machine', 'wash', 7),
    ('Помпа', 'machine', 'wash', 8),
    ('Электромотор', 'machine', 'wash', 9),
    ('Редуктор давления', 'machine', 'wash', 10),
    ('Общее', 'any', 'any', 11)
ON CONFLICT DO NOTHING;

-- Детали (нижний уровень) — привязка по имени группы
INSERT INTO component_types (group_id, name, sort_order)
SELECT g.id, d.name, d.ord
FROM (VALUES
    ('Пистолет моечный', 'Пистолет', 1),
    ('Пистолет моечный', 'Форсунка', 2),
    ('Пистолет моечный', 'Курок', 3),
    ('Пистолет моечный', 'Копьё', 4),
    ('Пистолет моечный', 'Форсункодержатель', 5),
    ('Пистолет химии', 'Пистолет химии', 1),
    ('Пистолет химии', 'Форсунка химии', 2),
    ('Пистолет химии', 'Прокладка химии', 3),
    ('Пистолет химии', 'Фильтр химии', 4),
    ('Бокс', 'Шланг АВД', 1),
    ('Бокс', 'Шланг химии', 2),
    ('Бокс', 'O-ring на шланге', 3),
    ('Бокс', 'Штуцеры шлангов в боксе', 4),
    ('Бокс', 'Магистральная труба', 5),
    ('Бокс', 'Светильник', 6),
    ('Аппаратная', 'Шланг в аппаратной', 1),
    ('Аппаратная', 'Пускатель в электрощитке', 2),
    ('Воздух', 'Клапан воздуха (электроклапан)', 1),
    ('Воздух', 'Шланг воздуха', 2),
    ('Воздух', 'Пистолет воздуха', 3),
    ('Пылесос', 'Кабель', 1),
    ('Пылесос', 'Кнопка', 2),
    ('Пылесос', 'Колёса', 3),
    ('Пылесос', 'Турбина (одна)', 4),
    ('Пылесос', 'Турбины (две)', 5),
    ('Пылесос', 'Турбины (три / все)', 6),
    ('Пылесос', 'Шланг пылесоса', 7),
    ('Регулятор давления', 'Регулятор давления', 1),
    ('Регулятор давления', 'Верхние прокладки', 2),
    ('Регулятор давления', 'Внутренние прокладки', 3),
    ('Регулятор давления', 'Прокладка клапана', 4),
    ('Регулятор давления', 'Тотал-стоп', 5),
    ('Регулятор давления', 'Штуцеры', 6),
    ('Регулятор давления', 'Клапан', 7),
    ('Помпа', 'Помпа', 1),
    ('Помпа', 'Головка помпы', 2),
    ('Помпа', 'Клапаны помпы', 3),
    ('Помпа', 'Плунжерные уплотнители', 4),
    ('Помпа', 'Подшипники', 5),
    ('Электромотор', 'Электромотор', 1),
    ('Электромотор', 'Подшипники электромотора', 2),
    ('Электромотор', 'Фланец электромотора', 3),
    ('Редуктор давления', 'Редуктор давления (в сборе)', 1),
    ('Общее', 'Замена масла', 1),
    ('Общее', 'Другое (комментарий)', 2)
) AS d(group_name, name, ord)
JOIN component_groups g ON g.name = d.group_name;

-- Получатели push-уведомлений: Максим (@PMN005) и Alex (@private_import)
INSERT INTO notification_recipients (name, chat_id, event_type, is_active) VALUES
    ('Максим', 5537830641, 'service_ticket', true),
    ('Alex',   188245803,  'service_ticket', true)
ON CONFLICT DO NOTHING;
