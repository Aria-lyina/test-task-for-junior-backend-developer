package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type RepeatTaskRepository struct {
	pool *pgxpool.Pool
}

func NewRepeatTaskRepository(pool *pgxpool.Pool) *RepeatTaskRepository {
	return &RepeatTaskRepository{pool: pool}
}

func (r *RepeatTaskRepository) Create(ctx context.Context, rt *taskdomain.RepeatTask) error {
	const query = `
		INSERT INTO repeat_tasks (
			title, description, status, period_id, rrule, custom_dates,
			enabled, last_generated_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		rt.Title, rt.Description, rt.Status,
		rt.PeriodID, rt.RRULE, rt.CustomDates,
		rt.Enabled, rt.LastGeneratedAt,
	)
	err := row.Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create repeat task: %w", err)
	}
	return nil
}

func (r *RepeatTaskRepository) GetByID(ctx context.Context, id int64) (*taskdomain.RepeatTask, error) {
	const query = `
		SELECT id, title, description, status, period_id, rrule, custom_dates,
		       enabled, last_generated_at, created_at, updated_at
		FROM repeat_tasks
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	rt, err := scanRepeatTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepeatTaskNotFound
		}
		return nil, err
	}
	return rt, nil
}

func (r *RepeatTaskRepository) Update(ctx context.Context, rt *taskdomain.RepeatTask) error {
	const query = `
		UPDATE repeat_tasks
		SET title = $1, description = $2, status = $3, period_id = $4,
		    rrule = $5, custom_dates = $6, enabled = $7, last_generated_at = $8,
		    updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		rt.Title, rt.Description, rt.Status,
		rt.PeriodID, rt.RRULE, rt.CustomDates,
		rt.Enabled, rt.LastGeneratedAt, rt.ID,
	)
	err := row.Scan(&rt.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRepeatTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update repeat task: %w", err)
	}
	return nil
}

func (r *RepeatTaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM repeat_tasks WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRepeatTaskNotFound
	}
	return nil
}

func (r *RepeatTaskRepository) ListActive(ctx context.Context) ([]taskdomain.RepeatTask, error) {
	const query = `
		SELECT id, title, description, status, period_id, rrule, custom_dates,
		       enabled, last_generated_at, created_at, updated_at
		FROM repeat_tasks
		WHERE enabled = true
		ORDER BY id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.RepeatTask
	for rows.Next() {
		rt, err := scanRepeatTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *rt)
	}
	return tasks, rows.Err()
}

func (r *RepeatTaskRepository) List(ctx context.Context) ([]taskdomain.RepeatTask, error) {
	const query = `
		SELECT id, title, description, status, period_id, rrule, custom_dates,
		       enabled, last_generated_at, created_at, updated_at
		FROM repeat_tasks
		ORDER BY id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.RepeatTask
	for rows.Next() {
		rt, err := scanRepeatTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *rt)
	}
	return tasks, rows.Err()
}


func (r *RepeatTaskRepository) UpdateLastGeneratedAt(ctx context.Context, id int64, ts time.Time) error {
	const query = `UPDATE repeat_tasks SET last_generated_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, ts, id)
	return err
}

func (r *RepeatTaskRepository) GetTasksWithRepeatInfo(ctx context.Context, limit, offset int) ([]taskdomain.TaskWithRepeatInfo, error) {
	const query = `
		SELECT
			t.id, t.title, t.description, t.status, t.repeat_task_id, t.created_at, t.updated_at,
			rt.id, rt.title, rt.description, rt.status, rt.period_id, rt.rrule, rt.custom_dates,
			rt.enabled, rt.last_generated_at, rt.created_at, rt.updated_at,
			p.id, p.code, p.title, p.rrule_template, p.created_at, p.updated_at
		FROM tasks t
		LEFT JOIN repeat_tasks rt ON t.repeat_task_id = rt.id
		LEFT JOIN periods p ON rt.period_id = p.id
		ORDER BY t.id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []taskdomain.TaskWithRepeatInfo
	for rows.Next() {
		info, err := scanTaskWithRepeatInfo(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, rows.Err()
}

// scanDateArray конвертирует pgtype.Array[pgtype.Date] в taskdomain.DateArray
func scanDateArray(src pgtype.Array[pgtype.Date]) taskdomain.DateArray {
	if len(src.Elements) == 0 {
		return taskdomain.DateArray{}
	}
	result := make(taskdomain.DateArray, 0, len(src.Elements))
	for _, el := range src.Elements {
		if !el.Valid {
			continue
		}
		t := time.Date(el.Time.Year(), el.Time.Month(), el.Time.Day(), 0, 0, 0, 0, time.UTC)
		result = append(result, t)
	}
	return result
}

// scanRepeatTask сканирует строку в RepeatTask.
// custom_dates читается через pgtype.Array[pgtype.Date] для поддержки бинарного протокола pgx.
func scanRepeatTask(scanner interface {
	Scan(dest ...any) error
}) (*taskdomain.RepeatTask, error) {
	var rt taskdomain.RepeatTask
	var periodID sql.NullInt64
	var rrule sql.NullString
	var lastGeneratedAt sql.NullTime
	var customDates pgtype.Array[pgtype.Date]

	err := scanner.Scan(
		&rt.ID,
		&rt.Title,
		&rt.Description,
		&rt.Status,
		&periodID,
		&rrule,
		&customDates,
		&rt.Enabled,
		&lastGeneratedAt,
		&rt.CreatedAt,
		&rt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan repeat task: %w", err)
	}

	if periodID.Valid {
		rt.PeriodID = &periodID.Int64
	}
	if rrule.Valid {
		rt.RRULE = &rrule.String
	}
	if lastGeneratedAt.Valid {
		rt.LastGeneratedAt = &lastGeneratedAt.Time
	}
	rt.CustomDates = scanDateArray(customDates)

	return &rt, nil
}

// scanTaskWithRepeatInfo сканирует строку JOIN задачи с шаблоном и периодом
func scanTaskWithRepeatInfo(scanner interface {
	Scan(dest ...any) error
}) (taskdomain.TaskWithRepeatInfo, error) {
	var info taskdomain.TaskWithRepeatInfo
	var repeatTaskID sql.NullInt64

	var rtID sql.NullInt64
	var rtTitle, rtDescription, rtStatus sql.NullString
	var rtPeriodID sql.NullInt64
	var rtRRULE sql.NullString
	var rtCustomDates pgtype.Array[pgtype.Date]
	var rtEnabled sql.NullBool
	var rtLastGeneratedAt, rtCreatedAt, rtUpdatedAt sql.NullTime

	var pID sql.NullInt64
	var pCode, pTitle, pRRULETemplate sql.NullString
	var pCreatedAt, pUpdatedAt sql.NullTime

	err := scanner.Scan(
		&info.Task.ID, &info.Task.Title, &info.Task.Description, &info.Task.Status,
		&repeatTaskID, &info.Task.CreatedAt, &info.Task.UpdatedAt,
		&rtID, &rtTitle, &rtDescription, &rtStatus,
		&rtPeriodID, &rtRRULE, &rtCustomDates,
		&rtEnabled, &rtLastGeneratedAt, &rtCreatedAt, &rtUpdatedAt,
		&pID, &pCode, &pTitle, &pRRULETemplate, &pCreatedAt, &pUpdatedAt,
	)
	if err != nil {
		return info, fmt.Errorf("failed to scan row: %w", err)
	}

	if repeatTaskID.Valid {
		info.Task.RepeatTaskID = &repeatTaskID.Int64
	}

	if rtID.Valid {
		rt := taskdomain.RepeatTask{
			ID:          rtID.Int64,
			Title:       rtTitle.String,
			Description: rtDescription.String,
			Status:      taskdomain.Status(rtStatus.String),
			Enabled:     rtEnabled.Bool,
			CustomDates: scanDateArray(rtCustomDates),
			CreatedAt:   rtCreatedAt.Time,
			UpdatedAt:   rtUpdatedAt.Time,
		}
		if rtPeriodID.Valid {
			rt.PeriodID = &rtPeriodID.Int64
		}
		if rtRRULE.Valid {
			rt.RRULE = &rtRRULE.String
		}
		if rtLastGeneratedAt.Valid {
			rt.LastGeneratedAt = &rtLastGeneratedAt.Time
		}
		info.RepeatTask = &rt
	}

	if pID.Valid {
		p := taskdomain.Period{
			ID:        pID.Int64,
			Code:      pCode.String,
			Title:     pTitle.String,
			CreatedAt: pCreatedAt.Time,
			UpdatedAt: pUpdatedAt.Time,
		}
		if pRRULETemplate.Valid {
			p.RRULETemplate = &pRRULETemplate.String
		}
		info.Period = &p
	}

	return info, nil
}

var ErrRepeatTaskNotFound = errors.New("repeat task not found")