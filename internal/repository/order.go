package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type OrderFilter struct {
	ShiftID   string
	CashierID string
	Status    string
	Page      int
	Limit     int
}

type OrderRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error)
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	AddItem(ctx context.Context, item *entity.OrderItem) error
	UpdateItem(ctx context.Context, item *entity.OrderItem) error
	DeleteItem(ctx context.Context, itemID string) error
	FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	// 1. Get Order
	queryOrder := `
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id, o.status,
		       o.subtotal, o.discount_amount, o.total, o.notes, o.created_at, o.updated_at,
		       u.id, u.name,
		       t.id, t.name
		FROM orders o
		JOIN users u ON u.id = o.cashier_id
		LEFT JOIN tables t ON t.id = o.table_id
		WHERE o.id = ?
	`

	row := r.db.QueryRowContext(ctx, queryOrder, id)

	var o entity.Order
	var notes sql.NullString
	var tID, tName sql.NullString
	var uID, uName string

	err := row.Scan(
		&o.ID,
		&o.ShiftID,
		&o.CashierID,
		&o.TableID,
		&o.PromoID,
		&o.Status,
		&o.Subtotal,
		&o.DiscountAmount,
		&o.Total,
		&notes,
		&o.CreatedAt,
		&o.UpdatedAt,
		&uID,
		&uName,
		&tID,
		&tName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if notes.Valid {
		o.Notes = notes.String
	}

	o.Cashier = &entity.User{
		ID:   uID,
		Name: uName,
	}

	if tID.Valid {
		o.Table = &entity.Table{
			ID:   tID.String,
			Name: tName.String,
		}
	}

	// 2. Get Items
	queryItems := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.subtotal,
		       oi.notes, oi.created_at, oi.updated_at,
		       p.id, p.name, p.price as product_price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = ?
	`

	rows, err := r.db.QueryContext(ctx, queryItems, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	o.Items = make([]entity.OrderItem, 0)
	for rows.Next() {
		var oi entity.OrderItem
		var itemNotes sql.NullString
		var pID, pName string
		var pPrice int64

		if err := rows.Scan(
			&oi.ID,
			&oi.OrderID,
			&oi.ProductID,
			&oi.Quantity,
			&oi.Price,
			&oi.Subtotal,
			&itemNotes,
			&oi.CreatedAt,
			&oi.UpdatedAt,
			&pID,
			&pName,
			&pPrice,
		); err != nil {
			return nil, err
		}

		if itemNotes.Valid {
			oi.Notes = itemNotes.String
		}

		oi.Product = &entity.Product{
			ID:    pID,
			Name:  pName,
			Price: pPrice,
		}

		o.Items = append(o.Items, oi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *orderRepository) FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error) {
	var conditions []string
	var args []interface{}

	if filter.ShiftID != "" {
		conditions = append(conditions, "o.shift_id = ?")
		args = append(args, filter.ShiftID)
	}
	if filter.CashierID != "" {
		conditions = append(conditions, "o.cashier_id = ?")
		args = append(args, filter.CashierID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(o.id) FROM orders o` + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	query := `
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id, o.status,
		       o.subtotal, o.discount_amount, o.total, o.notes, o.created_at, o.updated_at,
		       u.id, u.name
		FROM orders o
		JOIN users u ON u.id = o.cashier_id
	` + whereClause + `
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]entity.Order, 0)
	for rows.Next() {
		var o entity.Order
		var notes sql.NullString
		var uID, uName string

		if err := rows.Scan(
			&o.ID,
			&o.ShiftID,
			&o.CashierID,
			&o.TableID,
			&o.PromoID,
			&o.Status,
			&o.Subtotal,
			&o.DiscountAmount,
			&o.Total,
			&notes,
			&o.CreatedAt,
			&o.UpdatedAt,
			&uID,
			&uName,
		); err != nil {
			return nil, 0, err
		}

		if notes.Valid {
			o.Notes = notes.String
		}

		o.Cashier = &entity.User{
			ID:   uID,
			Name: uName,
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt

	var notes sql.NullString
	if order.Notes != "" {
		notes = sql.NullString{String: order.Notes, Valid: true}
	}

	query := `
		INSERT INTO orders (id, shift_id, cashier_id, table_id, promo_id, status, subtotal, discount_amount, total, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.ShiftID,
		order.CashierID,
		order.TableID, // driver automatically converts nil pointer to NULL
		order.PromoID,
		order.Status,
		order.Subtotal,
		order.DiscountAmount,
		order.Total,
		notes,
		order.CreatedAt,
		order.UpdatedAt,
	)

	return err
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	order.UpdatedAt = time.Now()

	var notes sql.NullString
	if order.Notes != "" {
		notes = sql.NullString{String: order.Notes, Valid: true}
	}

	query := `
		UPDATE orders
		SET table_id = ?, promo_id = ?, status = ?, subtotal = ?, discount_amount = ?, total = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.TableID,
		order.PromoID,
		order.Status,
		order.Subtotal,
		order.DiscountAmount,
		order.Total,
		notes,
		order.UpdatedAt,
		order.ID,
	)

	return err
}

func (r *orderRepository) AddItem(ctx context.Context, item *entity.OrderItem) error {
	item.ID = uuid.New().String()
	item.Subtotal = item.Price * int64(item.Quantity)
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt

	var notes sql.NullString
	if item.Notes != "" {
		notes = sql.NullString{String: item.Notes, Valid: true}
	}

	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price, subtotal, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.Price,
		item.Subtotal,
		notes,
		item.CreatedAt,
		item.UpdatedAt,
	)

	return err
}

func (r *orderRepository) UpdateItem(ctx context.Context, item *entity.OrderItem) error {
	item.UpdatedAt = time.Now()

	var notes sql.NullString
	if item.Notes != "" {
		notes = sql.NullString{String: item.Notes, Valid: true}
	}

	query := `
		UPDATE order_items
		SET quantity = ?, notes = ?, subtotal = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.Quantity,
		notes,
		item.Subtotal,
		item.UpdatedAt,
		item.ID,
	)

	return err
}

func (r *orderRepository) DeleteItem(ctx context.Context, itemID string) error {
	query := `DELETE FROM order_items WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, itemID)
	return err
}

func (r *orderRepository) FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price, subtotal, notes, created_at, updated_at
		FROM order_items
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, itemID)

	var oi entity.OrderItem
	var notes sql.NullString

	err := row.Scan(
		&oi.ID,
		&oi.OrderID,
		&oi.ProductID,
		&oi.Quantity,
		&oi.Price,
		&oi.Subtotal,
		&notes,
		&oi.CreatedAt,
		&oi.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if notes.Valid {
		oi.Notes = notes.String
	}

	return &oi, nil
}
