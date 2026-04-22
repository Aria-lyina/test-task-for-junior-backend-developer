package task

import (
	"database/sql/driver"
	"strings"
	"time"
)

// DateArray — тип для DATE[] из PostgreSQL.
// Запись в БД — через Value() в текстовом формате.
// Чтение из БД — репозиторий сканирует через pgtype.Array[pgtype.Date]
// и конвертирует в DateArray самостоятельно (см. repeat_task_repository.go).
type DateArray []time.Time

// Value реализует driver.Valuer — передаёт в PostgreSQL как {2026-01-01,2026-01-07}
func (d DateArray) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	if len(d) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(d))
	for i, t := range d {
		parts[i] = t.UTC().Format("2006-01-02")
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

type RepeatTask struct {
	ID int64 `json:"id"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`

	PeriodID    *int64    `json:"period_id,omitempty"`
	RRULE       *string   `json:"rrule,omitempty"`
	CustomDates DateArray `json:"custom_dates,omitempty"`

	Enabled         bool       `json:"enabled"`
	LastGeneratedAt *time.Time `json:"last_generated_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskWithRepeatInfo struct {
	Task       Task        `json:"task"`
	RepeatTask *RepeatTask `json:"repeat_task,omitempty"`
	Period     *Period     `json:"period,omitempty"`
}