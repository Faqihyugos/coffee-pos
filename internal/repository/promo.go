package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type PromoRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Promo, error)
	FindByCode(ctx context.Context, code string) (*entity.Promo, error)
	FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error)
	Create(ctx context.Context, promo *entity.Promo) error
	Update(ctx context.Context, promo *entity.Promo) error
	Delete(ctx context.Context, id string) error
	IncrementUsedCount(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) PromoRepository
}

type promoRepository struct {
	db sqlDB
}

func NewPromoRepository(db sqlDB) PromoRepository {
	return &promoRepository{db: db}
}

func (r *promoRepository) WithTx(tx *sql.Tx) PromoRepository {
	return &promoRepository{db: tx}
}

func (r *promoRepository) FindByID(ctx context.Context, id string) (*entity.Promo, error) {
	query := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit, 
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *promoRepository) FindByCode(ctx context.Context, code string) (*entity.Promo, error) {
	query := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit, 
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE code = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, code)
	return r.scanRow(row)
}

func (r *promoRepository) FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error) {
	countQuery := `SELECT COUNT(id) FROM promos WHERE deleted_at IS NULL`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit, 
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	promos := make([]entity.Promo, 0)
	for rows.Next() {
		var p entity.Promo
		var maxDiscount sql.NullInt64
		var usageLimit sql.NullInt64

		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Code,
			&p.Type,
			&p.Value,
			&p.MinOrder,
			&maxDiscount,
			&usageLimit,
			&p.UsedCount,
			&p.StartedAt,
			&p.EndedAt,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
		); err != nil {
			return nil, 0, err
		}

		if maxDiscount.Valid {
			p.MaxDiscount = &maxDiscount.Int64
		}
		if usageLimit.Valid {
			limitVal := int(usageLimit.Int64)
			p.UsageLimit = &limitVal
		}

		promos = append(promos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return promos, total, nil
}

func (r *promoRepository) Create(ctx context.Context, promo *entity.Promo) error {
	promo.ID = uuid.New().String()
	promo.CreatedAt = time.Now()
	promo.UpdatedAt = promo.CreatedAt

	var maxDiscount sql.NullInt64
	if promo.MaxDiscount != nil {
		maxDiscount = sql.NullInt64{Int64: *promo.MaxDiscount, Valid: true}
	}

	var usageLimit sql.NullInt64
	if promo.UsageLimit != nil {
		usageLimit = sql.NullInt64{Int64: int64(*promo.UsageLimit), Valid: true}
	}

	query := `
		INSERT INTO promos (id, name, code, type, value, min_order, max_discount, usage_limit, used_count, started_at, ended_at, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		promo.ID,
		promo.Name,
		promo.Code,
		promo.Type,
		promo.Value,
		promo.MinOrder,
		maxDiscount,
		usageLimit,
		promo.UsedCount,
		promo.StartedAt,
		promo.EndedAt,
		promo.IsActive,
		promo.CreatedAt,
		promo.UpdatedAt,
	)

	return err
}

func (r *promoRepository) Update(ctx context.Context, promo *entity.Promo) error {
	promo.UpdatedAt = time.Now()

	var maxDiscount sql.NullInt64
	if promo.MaxDiscount != nil {
		maxDiscount = sql.NullInt64{Int64: *promo.MaxDiscount, Valid: true}
	}

	var usageLimit sql.NullInt64
	if promo.UsageLimit != nil {
		usageLimit = sql.NullInt64{Int64: int64(*promo.UsageLimit), Valid: true}
	}

	query := `
		UPDATE promos
		SET name = ?, type = ?, value = ?, min_order = ?, max_discount = ?, usage_limit = ?,
		    started_at = ?, ended_at = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		promo.Name,
		promo.Type,
		promo.Value,
		promo.MinOrder,
		maxDiscount,
		usageLimit,
		promo.StartedAt,
		promo.EndedAt,
		promo.IsActive,
		promo.UpdatedAt,
		promo.ID,
	)

	return err
}

func (r *promoRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE promos
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

func (r *promoRepository) IncrementUsedCount(ctx context.Context, id string) error {
	query := `
		UPDATE promos
		SET used_count = used_count + 1, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *promoRepository) scanRow(row *sql.Row) (*entity.Promo, error) {
	var p entity.Promo
	var maxDiscount sql.NullInt64
	var usageLimit sql.NullInt64

	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Code,
		&p.Type,
		&p.Value,
		&p.MinOrder,
		&maxDiscount,
		&usageLimit,
		&p.UsedCount,
		&p.StartedAt,
		&p.EndedAt,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if maxDiscount.Valid {
		p.MaxDiscount = &maxDiscount.Int64
	}
	if usageLimit.Valid {
		limitVal := int(usageLimit.Int64)
		p.UsageLimit = &limitVal
	}

	return &p, nil
}
