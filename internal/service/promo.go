package service

import (
	"context"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
)

type PromoService struct {
	promoRepo repository.PromoRepository
}

func NewPromoService(promoRepo repository.PromoRepository) *PromoService {
	return &PromoService{
		promoRepo: promoRepo,
	}
}

func (s *PromoService) FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	return s.promoRepo.FindAll(ctx, page, limit)
}

func (s *PromoService) FindByID(ctx context.Context, id string) (*entity.Promo, error) {
	promo, err := s.promoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, errors.New("promo tidak ditemukan")
	}
	return promo, nil
}

func (s *PromoService) Create(ctx context.Context, req entity.CreatePromoRequest) (*entity.Promo, error) {
	existing, err := s.promoRepo.FindByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("kode promo sudah digunakan")
	}

	if req.Type == entity.PromoTypePercentage && req.Value > 100 {
		return nil, errors.New("nilai persentase tidak boleh lebih dari 100")
	}

	promo := &entity.Promo{
		Name:        req.Name,
		Code:        req.Code,
		Type:        req.Type,
		Value:       req.Value,
		MinOrder:    req.MinOrder,
		MaxDiscount: req.MaxDiscount,
		UsageLimit:  req.UsageLimit,
		StartedAt:   req.StartedAt,
		EndedAt:     req.EndedAt,
		IsActive:    true,
	}

	if err := s.promoRepo.Create(ctx, promo); err != nil {
		return nil, err
	}

	return promo, nil
}

func (s *PromoService) Update(ctx context.Context, id string, req entity.UpdatePromoRequest) (*entity.Promo, error) {
	promo, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		promo.Name = req.Name
	}
	if req.IsActive != nil {
		promo.IsActive = *req.IsActive
	}
	if req.StartedAt != nil {
		promo.StartedAt = *req.StartedAt
	}
	if req.EndedAt != nil {
		promo.EndedAt = *req.EndedAt
	}

	if err := s.promoRepo.Update(ctx, promo); err != nil {
		return nil, err
	}

	return promo, nil
}

func (s *PromoService) Delete(ctx context.Context, id string) error {
	_, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.promoRepo.Delete(ctx, id)
}
