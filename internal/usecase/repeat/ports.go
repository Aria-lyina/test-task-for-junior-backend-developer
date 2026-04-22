package repeat

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// PeriodRepository – интерфейс для работы с периодами
type PeriodRepository interface {
	Create(ctx context.Context, period *taskdomain.Period) error
	GetByID(ctx context.Context, id int64) (*taskdomain.Period, error)
	Update(ctx context.Context, period *taskdomain.Period) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Period, error)
}

// RepeatTaskRepository – интерфейс для работы с настройками повторения
type RepeatTaskRepository interface {
	Create(ctx context.Context, rt *taskdomain.RepeatTask) error
	GetByID(ctx context.Context, id int64) (*taskdomain.RepeatTask, error)
	Update(ctx context.Context, rt *taskdomain.RepeatTask) error
	Delete(ctx context.Context, id int64) error
	ListActive(ctx context.Context) ([]taskdomain.RepeatTask, error)
	List(ctx context.Context) ([]taskdomain.RepeatTask, error)
	UpdateLastGeneratedAt(ctx context.Context, id int64, ts time.Time) error
	GetTasksWithRepeatInfo(ctx context.Context, limit, offset int) ([]taskdomain.TaskWithRepeatInfo, error)
}

// TaskRepository – интерфейс для работы с задачами
type TaskRepository interface {
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Create(ctx context.Context, task *taskdomain.Task) error
	Update(ctx context.Context, task *taskdomain.Task) error
	Delete(ctx context.Context, id int64) error
}

// Usecase – полный интерфейс бизнес-логики
type Usecase interface {
	// Periods
	CreatePeriod(ctx context.Context, input CreatePeriodInput) (*taskdomain.Period, error)
	GetPeriodByID(ctx context.Context, id int64) (*taskdomain.Period, error)
	UpdatePeriod(ctx context.Context, id int64, input UpdatePeriodInput) (*taskdomain.Period, error)
	DeletePeriod(ctx context.Context, id int64) error
	ListPeriods(ctx context.Context) ([]taskdomain.Period, error)

	// RepeatTasks
	CreateRepeatTask(ctx context.Context, input CreateRepeatTaskInput) (*taskdomain.RepeatTask, error)
	GetRepeatTaskByID(ctx context.Context, id int64) (*taskdomain.RepeatTask, error)
	UpdateRepeatTask(ctx context.Context, id int64, input UpdateRepeatTaskInput) (*taskdomain.RepeatTask, error)
	DeleteRepeatTask(ctx context.Context, id int64) error
	ListActiveRepeatTasks(ctx context.Context) ([]taskdomain.RepeatTask, error)
	ListRepeatTasks(ctx context.Context) ([]taskdomain.RepeatTask, error)

	// Генерация
	GenerateTasksForRepeatTask(ctx context.Context, repeatTaskID int64, from, to time.Time) (int, error)
	GenerateAllTasks(ctx context.Context) (int, error)
	GetTasksWithRepeatInfo(ctx context.Context, limit, offset int) ([]taskdomain.TaskWithRepeatInfo, error)

	// Транзакционное создание задачи-шаблона с настройкой повторения
	CreateTaskWithRepeat(ctx context.Context, taskInput CreateTaskInput, repeatInput CreateRepeatTaskInput) (*taskdomain.Task, *taskdomain.RepeatTask, error)
}

// ========== DTO ==========

type CreatePeriodInput struct {
	Code          string
	Title         string
	RRULETemplate *string
}

type UpdatePeriodInput struct {
	Code          string
	Title         string
	RRULETemplate *string
}

type CreateRepeatTaskInput struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	// Ровно одно из трёх должно быть заполнено:
	// - PeriodID: стандартный период из справочника, rrule берётся автоматически
	// - RRULE: кастомное правило вручную
	// - CustomDates: конкретные даты
	PeriodID    *int64      `json:"period_id,omitempty"`
	RRULE       *string     `json:"rrule,omitempty"`
	CustomDates []time.Time `json:"custom_dates,omitempty"`
	Enabled     bool        `json:"enabled"`
}

type UpdateRepeatTaskInput struct {
	Title       *string            `json:"title,omitempty"`
	Description *string            `json:"description,omitempty"`
	Status      *taskdomain.Status `json:"status,omitempty"`
	// При обновлении правила повторения также должно быть заполнено ровно одно
	PeriodID    *int64      `json:"period_id,omitempty"`
	RRULE       *string     `json:"rrule,omitempty"`
	CustomDates []time.Time `json:"custom_dates,omitempty"`
	// Указатель, чтобы отличить "не передано" от явного false
	Enabled *bool `json:"enabled,omitempty"`
}

type CreateTaskInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}