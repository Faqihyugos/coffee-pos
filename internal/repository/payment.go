package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error)
	FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error
	UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error
	SaveRawNotification(ctx context.Context, id string, raw string) error
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) FindByID(ctx context.Context, id string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount, midtrans_order_id, midtrans_token, 
		       midtrans_url, raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount, midtrans_order_id, midtrans_token, 
		       midtrans_url, raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE order_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, orderID)
	return r.scanRow(row)
}

func (r *paymentRepository) FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount, midtrans_order_id, midtrans_token, 
		       midtrans_url, raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE midtrans_order_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, midtransOrderID)
	return r.scanRow(row)
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	payment.ID = uuid.New().String()
	payment.Status = entity.PaymentStatusPending
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = payment.CreatedAt

	var mOrderID, mToken, mURL, raw sql.NullString
	if payment.MidtransOrderID != nil {
		mOrderID = sql.NullString{String: *payment.MidtransOrderID, Valid: true}
	}
	if payment.MidtransToken != nil {
		mToken = sql.NullString{String: *payment.MidtransToken, Valid: true}
	}
	if payment.MidtransURL != nil {
		mURL = sql.NullString{String: *payment.MidtransURL, Valid: true}
	}
	if payment.RawNotification != nil {
		raw = sql.NullString{String: *payment.RawNotification, Valid: true}
	}

	var paidAt sql.NullTime
	if payment.PaidAt != nil {
		paidAt = sql.NullTime{Time: *payment.PaidAt, Valid: true}
	}

	query := `
		INSERT INTO payments (id, order_id, method, status, amount, midtrans_order_id, midtrans_token, midtrans_url, raw_notification, paid_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Method,
		payment.Status,
		payment.Amount,
		mOrderID,
		mToken,
		mURL,
		raw,
		paidAt,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	return err
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error {
	var paidAtNull sql.NullTime
	if paidAt != nil {
		paidAtNull = sql.NullTime{Time: *paidAt, Valid: true}
	}

	query := `
		UPDATE payments
		SET status = ?, paid_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, status, paidAtNull, time.Now(), id)
	return err
}

func (r *paymentRepository) UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error {
	query := `
		UPDATE payments
		SET midtrans_token = ?, midtrans_url = ?, midtrans_order_id = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, token, url, midtransOrderID, time.Now(), id)
	return err
}

func (r *paymentRepository) SaveRawNotification(ctx context.Context, id string, raw string) error {
	query := `
		UPDATE payments
		SET raw_notification = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, raw, time.Now(), id)
	return err
}

func (r *paymentRepository) scanRow(row *sql.Row) (*entity.Payment, error) {
	var p entity.Payment
	var mOrderID, mToken, mURL, raw sql.NullString
	var paidAt sql.NullTime

	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.Method,
		&p.Status,
		&p.Amount,
		&mOrderID,
		&mToken,
		&mURL,
		&raw,
		&paidAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if mOrderID.Valid {
		p.MidtransOrderID = &mOrderID.String
	}
	if mToken.Valid {
		p.MidtransToken = &mToken.String
	}
	if mURL.Valid {
		p.MidtransURL = &mURL.String
	}
	if raw.Valid {
		p.RawNotification = &raw.String
	}
	if paidAt.Valid {
		p.PaidAt = &paidAt.Time
	}

	return &p, nil
}
