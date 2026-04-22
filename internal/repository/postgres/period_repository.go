package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type PeriodRepository struct {
	pool *pgxpool.Pool
}

func NewPeriodRepository(pool *pgxpool.Pool) *PeriodRepository {
	return &PeriodRepository{pool: pool}
}

func (r *PeriodRepository) Create(ctx context.Context, p *taskdomain.Period) error {
	const query = `
		INSERT INTO periods (code, title, rrule_template)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query, p.Code, p.Title, p.RRULETemplate)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return err
}

func (r *PeriodRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Period, error) {
	const query = `
		SELECT id, code, title, rrule_template, created_at, updated_at
		FROM periods
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	p, err := scanPeriod(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrPeriodNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PeriodRepository) GetByCode(ctx context.Context, code string) (*taskdomain.Period, error) {
	const query = `
		SELECT id, code, title, rrule_template, created_at, updated_at
		FROM periods
		WHERE code = $1
	`
	row := r.pool.QueryRow(ctx, query, code)
	p, err := scanPeriod(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrPeriodNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PeriodRepository) Update(ctx context.Context, p *taskdomain.Period) error {
	const query = `
		UPDATE periods
		SET code = $1, title = $2, rrule_template = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	row := r.pool.QueryRow(ctx, query, p.Code, p.Title, p.RRULETemplate, p.ID)
	err := row.Scan(&p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return taskdomain.ErrPeriodNotFound
	}
	return err
}

func (r *PeriodRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM periods WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrPeriodNotFound
	}
	return nil
}

func (r *PeriodRepository) List(ctx context.Context) ([]taskdomain.Period, error) {
	const query = `
		SELECT id, code, title, rrule_template, created_at, updated_at
		FROM periods
		ORDER BY id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []taskdomain.Period
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, err
		}
		periods = append(periods, *p)
	}
	return periods, rows.Err()
}

// scanPeriod сканирует строку в Period с учётом NULL для rrule_template
func scanPeriod(scanner interface {
	Scan(dest ...any) error
}) (*taskdomain.Period, error) {
	var p taskdomain.Period
	var rruleTemplate sql.NullString

	err := scanner.Scan(
		&p.ID,
		&p.Code,
		&p.Title,
		&rruleTemplate,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if rruleTemplate.Valid {
		p.RRULETemplate = &rruleTemplate.String
	} else {
		p.RRULETemplate = nil
	}
	return &p, nil
}