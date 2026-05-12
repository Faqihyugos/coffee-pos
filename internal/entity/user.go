package entity

import "time"

const (
	RoleOwner   = "owner"
	RoleCashier = "cashier"
)

type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

func (u *User) IsCashier() bool {
	return u.Role == RoleCashier
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Role     string `json:"role"     validate:"required,oneof=owner cashier"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

type CreateCashierRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type UpdateCashierRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}
