package entity

import "time"

const (
	PromoTypePercentage = "percentage"
	PromoTypeFixed      = "fixed"
)

type Promo struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Type         string     `json:"type"`
	Value        int64      `json:"value"`
	MinOrder     int64      `json:"min_order"`
	MaxDiscount  *int64     `json:"max_discount,omitempty"`
	UsageLimit   *int       `json:"usage_limit,omitempty"`
	UsedCount    int        `json:"used_count"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      time.Time  `json:"ended_at"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

func (p *Promo) IsValid() bool {
	if !p.IsActive {
		return false
	}
	now := time.Now()
	if now.Before(p.StartedAt) || now.After(p.EndedAt) {
		return false
	}
	if p.UsageLimit != nil && p.UsedCount >= *p.UsageLimit {
		return false
	}
	return true
}

type CreatePromoRequest struct {
	Name        string    `json:"name"         validate:"required,min=2,max=200"`
	Code        string    `json:"code"         validate:"required,min=3,max=50"`
	Type        string    `json:"type"         validate:"required,oneof=percentage fixed"`
	Value       int64     `json:"value"        validate:"required,min=1"`
	MinOrder    int64     `json:"min_order"    validate:"min=0"`
	MaxDiscount *int64    `json:"max_discount" validate:"omitempty,min=1"`
	UsageLimit  *int      `json:"usage_limit"  validate:"omitempty,min=1"`
	StartedAt   time.Time `json:"started_at"   validate:"required"`
	EndedAt     time.Time `json:"ended_at"     validate:"required"`
}

type UpdatePromoRequest struct {
	Name      string     `json:"name"       validate:"omitempty,min=2,max=200"`
	IsActive  *bool      `json:"is_active"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
}
