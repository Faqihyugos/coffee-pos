package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type TableRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Table, error)
	FindAll(ctx context.Context) ([]entity.Table, error)
	Create(ctx context.Context, table *entity.Table) error
	Update(ctx context.Context, table *entity.Table) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type tableRepository struct {
	db *sql.DB
}

func NewTableRepository(db *sql.DB) TableRepository {
	return &tableRepository{db: db}
}

func (r *tableRepository) FindByID(ctx context.Context, id string) (*entity.Table, error) {
	query := `
		SELECT id, name, capacity, status, created_at, updated_at, deleted_at
		FROM tables
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *tableRepository) FindAll(ctx context.Context) ([]entity.Table, error) {
	query := `
		SELECT id, name, capacity, status, created_at, updated_at, deleted_at
		FROM tables
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]entity.Table, 0)
	for rows.Next() {
		var table entity.Table
		if err := rows.Scan(
			&table.ID,
			&table.Name,
			&table.Capacity,
			&table.Status,
			&table.CreatedAt,
			&table.UpdatedAt,
			&table.DeletedAt,
		); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *tableRepository) Create(ctx context.Context, table *entity.Table) error {
	table.ID = uuid.New().String()
	table.CreatedAt = time.Now()
	table.UpdatedAt = table.CreatedAt

	query := `
		INSERT INTO tables (id, name, capacity, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		table.ID,
		table.Name,
		table.Capacity,
		table.Status,
		table.CreatedAt,
		table.UpdatedAt,
	)

	return err
}

func (r *tableRepository) Update(ctx context.Context, table *entity.Table) error {
	table.UpdatedAt = time.Now()

	query := `
		UPDATE tables
		SET name = ?, capacity = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		table.Name,
		table.Capacity,
		table.UpdatedAt,
		table.ID,
	)

	return err
}

func (r *tableRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE tables
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

func (r *tableRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE tables
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, now, id)
	return err
}

func (r *tableRepository) scanRow(row *sql.Row) (*entity.Table, error) {
	var table entity.Table
	err := row.Scan(
		&table.ID,
		&table.Name,
		&table.Capacity,
		&table.Status,
		&table.CreatedAt,
		&table.UpdatedAt,
		&table.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &table, nil
}
