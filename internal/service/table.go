package service

import (
	"context"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
)

type TableService struct {
	tableRepo repository.TableRepository
}

func NewTableService(tableRepo repository.TableRepository) *TableService {
	return &TableService{
		tableRepo: tableRepo,
	}
}

func (s *TableService) FindAll(ctx context.Context) ([]entity.Table, error) {
	return s.tableRepo.FindAll(ctx)
}

func (s *TableService) FindByID(ctx context.Context, id string) (*entity.Table, error) {
	table, err := s.tableRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, errors.New("meja tidak ditemukan")
	}
	return table, nil
}

func (s *TableService) Create(ctx context.Context, req entity.CreateTableRequest) (*entity.Table, error) {
	existing, err := s.tableRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("nama meja sudah digunakan")
	}

	table := &entity.Table{
		Name:     req.Name,
		Capacity: req.Capacity,
		Status:   entity.TableStatusAvailable,
	}

	if err := s.tableRepo.Create(ctx, table); err != nil {
		return nil, err
	}

	return table, nil
}

func (s *TableService) Update(ctx context.Context, id string, req entity.UpdateTableRequest) (*entity.Table, error) {
	table, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	isUpdated := false

	if req.Name != "" {
		existing, err := s.tableRepo.FindByName(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("nama meja sudah digunakan")
		}

		if table.Name != req.Name {
			table.Name = req.Name
			isUpdated = true
		}
	}

	if req.Capacity > 0 {
		if table.Capacity != req.Capacity {
			table.Capacity = req.Capacity
			isUpdated = true
		}
	}

	if isUpdated {
		if err := s.tableRepo.Update(ctx, table); err != nil {
			return nil, err
		}
	}

	if req.Status != "" {
		if err := s.tableRepo.UpdateStatus(ctx, id, req.Status); err != nil {
			return nil, err
		}
		table.Status = req.Status
	}

	return table, nil
}

func (s *TableService) Delete(ctx context.Context, id string) error {
	_, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.tableRepo.Delete(ctx, id)
}
