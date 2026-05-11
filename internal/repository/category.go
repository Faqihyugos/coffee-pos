package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindByName(ctx context.Context, name string) (*entity.Category, error)
	FindAll(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) CategoryRepository
}

type categoryRepository struct {
	db sqlDB
}

func NewCategoryRepository(db sqlDB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) WithTx(tx *sql.Tx) CategoryRepository {
	return &categoryRepository{db: tx}
}

func (r *categoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *categoryRepository) FindByName(ctx context.Context, name string) (*entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE name = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, name)
	return r.scanRow(row)
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]entity.Category, 0)
	for rows.Next() {
		var category entity.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	category.ID = uuid.New().String()
	category.CreatedAt = time.Now()
	category.UpdatedAt = category.CreatedAt

	query := `
		INSERT INTO categories (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.ID,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)

	return err
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	category.UpdatedAt = time.Now()

	query := `
		UPDATE categories
		SET name = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.Name,
		category.UpdatedAt,
		category.ID,
	)

	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE categories
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

func (r *categoryRepository) scanRow(row *sql.Row) (*entity.Category, error) {
	var category entity.Category
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}
