package service

import (
	"context"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) FindAll(ctx context.Context) ([]entity.Category, error) {
	return s.categoryRepo.FindAll(ctx)
}

func (s *CategoryService) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("kategori tidak ditemukan")
	}
	return category, nil
}

func (s *CategoryService) Create(ctx context.Context, req entity.CreateCategoryRequest) (*entity.Category, error) {
	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("nama kategori sudah digunakan")
	}

	category := &entity.Category{
		Name: req.Name,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id string, req entity.UpdateCategoryRequest) (*entity.Category, error) {
	category, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, errors.New("nama kategori sudah digunakan")
	}

	category.Name = req.Name

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	_, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.categoryRepo.Delete(ctx, id)
}
