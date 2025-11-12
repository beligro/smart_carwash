-- Добавление таблиц для управления статусом мойки (открыта/закрыта)

-- Таблица текущего статуса мойки (singleton - всегда одна запись)
CREATE TABLE IF NOT EXISTS carwash_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_closed BOOLEAN DEFAULT false NOT NULL,
    closed_reason TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Таблица истории изменений статуса мойки
CREATE TABLE IF NOT EXISTS carwash_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_closed BOOLEAN NOT NULL,
    closed_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    created_by UUID
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_carwash_status_is_closed ON carwash_status(is_closed);
CREATE INDEX IF NOT EXISTS idx_carwash_status_history_created_at ON carwash_status_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_carwash_status_history_created_by ON carwash_status_history(created_by);

-- Вставляем начальную запись со статусом "открыта" (только если таблица пустая)
INSERT INTO carwash_status (is_closed, closed_reason, updated_at, created_at)
SELECT false, NULL, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM carwash_status);

