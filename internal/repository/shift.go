package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/google/uuid"
)

type ShiftRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Shift, error)
	FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error)
	FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error)
	Create(ctx context.Context, shift *entity.Shift) error
	Close(ctx context.Context, id string, closingCash int64, notes string) error
	WithTx(tx *sql.Tx) ShiftRepository
}

type shiftRepository struct {
	db sqlDB
}

func NewShiftRepository(db sqlDB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) WithTx(tx *sql.Tx) ShiftRepository {
	return &shiftRepository{db: tx}
}

func (r *shiftRepository) FindByID(ctx context.Context, id string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash, 
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRowWithUser(row)
}

func (r *shiftRepository) FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash, 
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.cashier_id = ? AND s.status = 'open'
	`

	row := r.db.QueryRowContext(ctx, query, cashierID)
	return r.scanRowWithUser(row)
}

func (r *shiftRepository) FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error) {
	var args []interface{}
	whereClause := ""

	if cashierID != "" {
		whereClause = "WHERE s.cashier_id = ?"
		args = append(args, cashierID)
	}

	// 1. Get Total Count
	countQuery := `SELECT COUNT(s.id) FROM shifts s ` + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 2. Setup Pagination
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 3. Get Data
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash, 
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
	` + whereClause + `
		ORDER BY s.opened_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	shifts := make([]entity.Shift, 0)
	for rows.Next() {
		var s entity.Shift
		var notes sql.NullString
		var closedAt sql.NullTime
		var closingCash sql.NullInt64
		var uID, uName string

		if err := rows.Scan(
			&s.ID,
			&s.CashierID,
			&s.OpenedAt,
			&closedAt,
			&s.OpeningCash,
			&closingCash,
			&s.TotalTransactions,
			&s.Status,
			&notes,
			&s.CreatedAt,
			&s.UpdatedAt,
			&uID,
			&uName,
		); err != nil {
			return nil, 0, err
		}

		if notes.Valid {
			s.Notes = notes.String
		}
		if closedAt.Valid {
			s.ClosedAt = &closedAt.Time
		}
		if closingCash.Valid {
			cCash := closingCash.Int64
			s.ClosingCash = &cCash
		}

		s.Cashier = &entity.User{
			ID:   uID,
			Name: uName,
		}

		shifts = append(shifts, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return shifts, total, nil
}

func (r *shiftRepository) Create(ctx context.Context, shift *entity.Shift) error {
	shift.ID = uuid.New().String()
	now := time.Now()
	shift.OpenedAt = now
	shift.CreatedAt = now
	shift.UpdatedAt = now
	shift.Status = entity.ShiftStatusOpen

	var notes sql.NullString
	if shift.Notes != "" {
		notes = sql.NullString{String: shift.Notes, Valid: true}
	}

	query := `
		INSERT INTO shifts (id, cashier_id, opened_at, opening_cash, total_transactions, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		shift.ID,
		shift.CashierID,
		shift.OpenedAt,
		shift.OpeningCash,
		shift.TotalTransactions,
		shift.Status,
		notes,
		shift.CreatedAt,
		shift.UpdatedAt,
	)

	return err
}

func (r *shiftRepository) Close(ctx context.Context, id string, closingCash int64, notes string) error {
	var notesNull sql.NullString
	if notes != "" {
		notesNull = sql.NullString{String: notes, Valid: true}
	}

	query := `
		UPDATE shifts
		SET status = ?, closed_at = ?, closing_cash = ?, notes = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, entity.ShiftStatusClosed, now, closingCash, notesNull, now, id, entity.ShiftStatusOpen)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("shift tidak ditemukan atau sudah ditutup")
	}

	return nil
}

func (r *shiftRepository) scanRowWithUser(row *sql.Row) (*entity.Shift, error) {
	var s entity.Shift
	var notes sql.NullString
	var closedAt sql.NullTime
	var closingCash sql.NullInt64
	var uID, uName string

	err := row.Scan(
		&s.ID,
		&s.CashierID,
		&s.OpenedAt,
		&closedAt,
		&s.OpeningCash,
		&closingCash,
		&s.TotalTransactions,
		&s.Status,
		&notes,
		&s.CreatedAt,
		&s.UpdatedAt,
		&uID,
		&uName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if notes.Valid {
		s.Notes = notes.String
	}
	if closedAt.Valid {
		s.ClosedAt = &closedAt.Time
	}
	if closingCash.Valid {
		cCash := closingCash.Int64
		s.ClosingCash = &cCash
	}

	s.Cashier = &entity.User{
		ID:   uID,
		Name: uName,
	}

	return &s, nil
}
