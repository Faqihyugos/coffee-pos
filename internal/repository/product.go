package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type ProductFilter struct {
	CategoryID string
	IsActive   *bool
	Search     string
	Page       int
	Limit      int
}

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error)
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, stock int) error
	WithTx(tx *sql.Tx) ProductRepository
}

type productRepository struct {
	db sqlDB
}

func NewProductRepository(db sqlDB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) WithTx(tx *sql.Tx) ProductRepository {
	return &productRepository{db: tx}
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE p.id = ? AND p.deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var p entity.Product
	var cID, cName sql.NullString
	var desc, img sql.NullString

	err := row.Scan(
		&p.ID,
		&p.CategoryID,
		&p.Name,
		&desc,
		&p.Price,
		&p.Stock,
		&img,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
		&cID,
		&cName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if desc.Valid {
		p.Description = desc.String
	}
	if img.Valid {
		p.ImageURL = img.String
	}

	if cID.Valid {
		p.Category = &entity.Category{
			ID:   cID.String,
			Name: cName.String,
		}
	}

	return &p, nil
}

func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error) {
	var conditions []string
	var args []interface{}

	if filter.CategoryID != "" {
		conditions = append(conditions, "p.category_id = ?")
		args = append(args, filter.CategoryID)
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "p.is_active = ?")
		args = append(args, *filter.IsActive)
	}

	if filter.Search != "" {
		conditions = append(conditions, "p.name LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// 1. Get Total Count
	countQuery := `SELECT COUNT(p.id) FROM products p WHERE p.deleted_at IS NULL` + whereClause
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
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE p.deleted_at IS NULL` + whereClause + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]entity.Product, 0)
	for rows.Next() {
		var p entity.Product
		var cID, cName sql.NullString
		var desc, img sql.NullString

		if err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&desc,
			&p.Price,
			&p.Stock,
			&img,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
			&cID,
			&cName,
		); err != nil {
			return nil, 0, err
		}

		if desc.Valid {
			p.Description = desc.String
		}
		if img.Valid {
			p.ImageURL = img.String
		}

		if cID.Valid {
			p.Category = &entity.Category{
				ID:   cID.String,
				Name: cName.String,
			}
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = product.CreatedAt

	var desc, img sql.NullString
	if product.Description != "" {
		desc = sql.NullString{String: product.Description, Valid: true}
	}
	if product.ImageURL != "" {
		img = sql.NullString{String: product.ImageURL, Valid: true}
	}

	query := `
		INSERT INTO products (id, category_id, name, description, price, stock, image_url, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.ID,
		product.CategoryID,
		product.Name,
		desc,
		product.Price,
		product.Stock,
		img,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	)

	return err
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()

	var desc, img sql.NullString
	if product.Description != "" {
		desc = sql.NullString{String: product.Description, Valid: true}
	}
	if product.ImageURL != "" {
		img = sql.NullString{String: product.ImageURL, Valid: true}
	}

	query := `
		UPDATE products
		SET category_id = ?, name = ?, description = ?, price = ?, image_url = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.CategoryID,
		product.Name,
		desc,
		product.Price,
		img,
		product.IsActive,
		product.UpdatedAt,
		product.ID,
	)

	return err
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE products
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

func (r *productRepository) UpdateStock(ctx context.Context, id string, stock int) error {
	query := `
		UPDATE products
		SET stock = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, stock, time.Now(), id)
	return err
}
