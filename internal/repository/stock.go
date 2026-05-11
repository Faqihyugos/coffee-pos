package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type StockFilter struct {
	ProductID string
	UserID    string
	Type      string
	Page      int
	Limit     int
}

type StockRepository interface {
	Create(ctx context.Context, movement *entity.StockMovement) error
	FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error)
	FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error)
	WithTx(tx *sql.Tx) StockRepository
}

type stockRepository struct {
	db sqlDB
}

func NewStockRepository(db sqlDB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) WithTx(tx *sql.Tx) StockRepository {
	return &stockRepository{db: tx}
}

func (r *stockRepository) Create(ctx context.Context, movement *entity.StockMovement) error {
	movement.ID = uuid.New().String()
	movement.CreatedAt = time.Now()
	movement.UpdatedAt = movement.CreatedAt

	var notes sql.NullString
	if movement.Notes != "" {
		notes = sql.NullString{String: movement.Notes, Valid: true}
	}

	query := `
		INSERT INTO stock_movements (id, product_id, user_id, type, quantity, stock_before, stock_after, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		movement.ID,
		movement.ProductID,
		movement.UserID,
		movement.Type,
		movement.Quantity,
		movement.StockBefore,
		movement.StockAfter,
		notes,
		movement.CreatedAt,
		movement.UpdatedAt,
	)

	return err
}

func (r *stockRepository) FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error) {
	filter.ProductID = productID
	return r.FindAll(ctx, filter)
}

func (r *stockRepository) FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error) {
	var conditions []string
	var args []interface{}

	if filter.ProductID != "" {
		conditions = append(conditions, "sm.product_id = ?")
		args = append(args, filter.ProductID)
	}

	if filter.UserID != "" {
		conditions = append(conditions, "sm.user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Type != "" {
		conditions = append(conditions, "sm.type = ?")
		args = append(args, filter.Type)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Get Total Count
	countQuery := `
		SELECT COUNT(sm.id)
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		JOIN users u ON u.id = sm.user_id
	` + whereClause

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 2. Setup Pagination
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	// 3. Get Data
	query := `
		SELECT sm.id, sm.product_id, sm.user_id, sm.type, sm.quantity,
		       sm.stock_before, sm.stock_after, sm.notes, sm.created_at, sm.updated_at,
		       p.id, p.name,
		       u.id, u.name
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		JOIN users u ON u.id = sm.user_id
	` + whereClause + `
		ORDER BY sm.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movements := make([]entity.StockMovement, 0)
	for rows.Next() {
		var sm entity.StockMovement
		var notes sql.NullString
		var pID, pName string
		var uID, uName string

		if err := rows.Scan(
			&sm.ID,
			&sm.ProductID,
			&sm.UserID,
			&sm.Type,
			&sm.Quantity,
			&sm.StockBefore,
			&sm.StockAfter,
			&notes,
			&sm.CreatedAt,
			&sm.UpdatedAt,
			&pID,
			&pName,
			&uID,
			&uName,
		); err != nil {
			return nil, 0, err
		}

		if notes.Valid {
			sm.Notes = notes.String
		}

		sm.Product = &entity.Product{
			ID:   pID,
			Name: pName,
		}

		sm.User = &entity.User{
			ID:   uID,
			Name: uName,
		}

		movements = append(movements, sm)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}
