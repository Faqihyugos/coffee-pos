package service

import (
	"context"
	"errors"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
)

type UserService struct {
	userRepo    repository.UserRepository
	authService *AuthService
}

func NewUserService(userRepo repository.UserRepository, authService *AuthService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *UserService) FindAllCashiers(ctx context.Context) ([]entity.UserResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var cashiers []entity.UserResponse
	for _, user := range users {
		if user.Role == entity.RoleCashier {
			cashiers = append(cashiers, user.ToResponse())
		}
	}

	if cashiers == nil {
		cashiers = []entity.UserResponse{}
	}

	return cashiers, nil
}

func (s *UserService) CreateCashier(ctx context.Context, req entity.CreateCashierRequest) (*entity.UserResponse, error) {
	registerReq := entity.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     entity.RoleCashier,
	}

	return s.authService.Register(ctx, registerReq)
}

func (s *UserService) Update(ctx context.Context, id string, req entity.UpdateCashierRequest) (*entity.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("cashier tidak ditemukan")
	}

	if user.Role != entity.RoleCashier {
		return nil, errors.New("user bukan cashier")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *UserService) ToggleStatus(ctx context.Context, id string) (*entity.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("cashier tidak ditemukan")
	}

	if user.Role != entity.RoleCashier {
		return nil, errors.New("user bukan cashier")
	}

	user.IsActive = !user.IsActive

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}
