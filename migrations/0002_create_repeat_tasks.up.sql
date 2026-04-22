CREATE TABLE IF NOT EXISTS periods (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,          -- Уникальный код, например 'daily', 'weekdays'
    title TEXT NOT NULL,                -- Человеческое название "Ежедневно"
    rrule_template TEXT,                -- Базовый RRULE (может быть NULL)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS repeat_tasks (
    id BIGSERIAL PRIMARY KEY,
    
    -- Шаблон для генерируемых задач
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'new',
    
    -- Настройки повторения
    period_id BIGINT REFERENCES periods(id) ON DELETE SET NULL,
    rrule TEXT,
    custom_dates DATE[],
    
    -- Управление
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_generated_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

    CONSTRAINT chk_repeat_rule CHECK (
        period_id IS NOT NULL OR rrule IS NOT NULL OR custom_dates IS NOT NULL
    )

);

-- Индексы
CREATE INDEX idx_repeat_tasks_enabled ON repeat_tasks(enabled);
CREATE INDEX idx_repeat_tasks_period_id ON repeat_tasks(period_id);

ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT REFERENCES repeat_tasks(id) ON DELETE SET NULL;
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);