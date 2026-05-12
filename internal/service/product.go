package service

import (
	"context"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
)

type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *ProductService) FindAll(ctx context.Context, filter repository.ProductFilter) ([]entity.Product, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	return s.productRepo.FindAll(ctx, filter)
}

func (s *ProductService) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("produk tidak ditemukan")
	}
	return product, nil
}

func (s *ProductService) Create(ctx context.Context, req entity.CreateProductRequest) (*entity.Product, error) {
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("kategori tidak ditemukan")
	}

	product := &entity.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Update(ctx context.Context, id string, req entity.UpdateProductRequest) (*entity.Product, error) {
	product, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.CategoryID != "" {
		category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, errors.New("kategori tidak ditemukan")
		}
		product.CategoryID = req.CategoryID
	}

	if req.Name != "" {
		product.Name = req.Name
	}

	if req.Description != "" {
		product.Description = req.Description
	}

	if req.Price > 0 {
		product.Price = req.Price
	}

	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	_, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.productRepo.Delete(ctx, id)
}
