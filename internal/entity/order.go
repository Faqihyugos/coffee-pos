package entity

import "time"

const (
	OrderStatusDraft          = "draft"
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusCancelled      = "cancelled"
)

type OrderItem struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
	Subtotal  int64     `json:"subtotal"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID             string      `json:"id"`
	ShiftID        string      `json:"shift_id"`
	CashierID      string      `json:"cashier_id"`
	Cashier        *User       `json:"cashier,omitempty"`
	TableID        *string     `json:"table_id,omitempty"`
	Table          *Table      `json:"table,omitempty"`
	PromoID        *string     `json:"promo_id,omitempty"`
	Promo          *Promo      `json:"promo,omitempty"`
	Items          []OrderItem `json:"items,omitempty"`
	Status         string      `json:"status"`
	Subtotal       int64       `json:"subtotal"`
	DiscountAmount int64       `json:"discount_amount"`
	Total          int64       `json:"total"`
	Notes          string      `json:"notes,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type CreateOrderRequest struct {
	TableID string `json:"table_id" validate:"omitempty,uuid4"`
	Notes   string `json:"notes"    validate:"max=500"`
}

type AddOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid4"`
	Quantity  int    `json:"quantity"   validate:"required,min=1,max=100"`
	Notes     string `json:"notes"      validate:"max=200"`
}

type UpdateOrderItemRequest struct {
	Quantity int    `json:"quantity" validate:"required,min=1,max=100"`
	Notes    string `json:"notes"    validate:"max=200"`
}

type ApplyPromoRequest struct {
	PromoCode string `json:"promo_code" validate:"required,min=3,max=50"`
}
