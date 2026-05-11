package entity

import "time"

type Category struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
