package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, t *taskdomain.Task) error {
	const query = `
		INSERT INTO tasks (title, description, status, parent_task_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		t.Title, t.Description, t.Status, t.RepeatTaskID, t.CreatedAt, t.UpdatedAt,
	)
	err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, parent_task_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}
	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *taskdomain.Task) error {
	const query = `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, parent_task_id = $4, updated_at = $5
		WHERE id = $6
		RETURNING updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		t.Title, t.Description, t.Status, t.RepeatTaskID, t.UpdatedAt, t.ID,
	)
	err := row.Scan(&t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return taskdomain.ErrNotFound
	}
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, parent_task_id, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

// scanTask сканирует строку в Task, правильно обрабатывая NULL для parent_task_id
func scanTask(scanner interface {
	Scan(dest ...any) error
}) (*taskdomain.Task, error) {
	var t taskdomain.Task
	var parentID sql.NullInt64

	err := scanner.Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&parentID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		t.RepeatTaskID = &parentID.Int64
	} else {
		t.RepeatTaskID = nil
	}
	return &t, nil
}