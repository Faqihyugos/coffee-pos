package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
	"github.com/faqihyugos/coffee-pos/pkg/txmanager"
)

type StockService struct {
	stockRepo   repository.StockRepository
	productRepo repository.ProductRepository
	txManager   *txmanager.TxManager
}

func NewStockService(
	stockRepo repository.StockRepository,
	productRepo repository.ProductRepository,
	txManager *txmanager.TxManager,
) *StockService {
	return &StockService{
		stockRepo:   stockRepo,
		productRepo: productRepo,
		txManager:   txManager,
	}
}

func (s *StockService) GetStock(ctx context.Context, productID string) (*entity.Product, error) {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("produk tidak ditemukan")
	}
	return product, nil
}

func (s *StockService) Adjust(ctx context.Context, productID string, userID string, req entity.StockAdjustmentRequest) error {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("produk tidak ditemukan")
	}

	var movementQty int
	var newStock int

	if req.Type == "in" || req.Type == "adjustment" {
		movementQty = req.Quantity
		newStock = product.Stock + req.Quantity
	} else if req.Type == "out" {
		movementQty = -req.Quantity
		newStock = product.Stock - req.Quantity
	} else {
		return errors.New("tipe pergerakan stok tidak valid")
	}

	if newStock < 0 {
		return errors.New("stok tidak cukup")
	}

	return s.txManager.WithTx(ctx, func(tx *sql.Tx) error {
		// 1. Buat movement dengan txStockRepo
		txStockRepo := s.stockRepo.WithTx(tx)
		
		movement := &entity.StockMovement{
			ProductID:   productID,
			UserID:      userID,
			Type:        req.Type,
			Quantity:    movementQty,
			StockBefore: product.Stock,
			StockAfter:  newStock,
			Notes:       req.Notes,
		}

		if err := txStockRepo.Create(ctx, movement); err != nil {
			return err
		}

		// 2. Update stock di products table dengan txProductRepo
		txProductRepo := s.productRepo.WithTx(tx)
		if err := txProductRepo.UpdateStock(ctx, productID, newStock); err != nil {
			return err
		}

		return nil
	})
}
