package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindAll(ctx context.Context) ([]entity.User, error)
	WithTx(tx *sql.Tx) UserRepository
}

type userRepository struct {
	db sqlDB
}

func NewUserRepository(db sqlDB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTx(tx *sql.Tx) UserRepository {
	return &userRepository{db: tx}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanRow(row)
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	query := `
		INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET name = ?, email = ?, role = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

func (r *userRepository) FindAll(ctx context.Context) ([]entity.User, error) {
	query := `
		SELECT id, name, email, password, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entity.User, 0)
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) scanRow(row *sql.Row) (*entity.User, error) {
	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil when user is not found
		}
		return nil, err
	}

	return &user, nil
}
