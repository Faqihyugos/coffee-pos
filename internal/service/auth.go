package service

import (
	"context"
	"errors"
	"time"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpireHours int
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtExpireHours int,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

func (s *AuthService) Register(ctx context.Context, req entity.RegisterRequest) (*entity.UserResponse, error) {
	// 1. Cek email apakah sudah dipakai
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// 2. Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	// 3. Buat entity.User baru
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		IsActive: true, // 4. Set IsActive = true
	}

	// 5. Simpan ke database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 6. Return response
	resp := user.ToResponse()
	return &resp, nil
}

func (s *AuthService) Login(ctx context.Context, req entity.LoginRequest) (*entity.LoginResponse, error) {
	// 1. Cari user berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("email atau password salah")
	}

	// 2. Cek apakah user aktif
	if !user.IsActive {
		return nil, errors.New("akun dinonaktifkan")
	}

	// 3. Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	// 4. Generate JWT token
	expiredAt := time.Now().Add(time.Duration(s.jwtExpireHours) * time.Hour)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expiredAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	// 5. Return LoginResponse
	return &entity.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiredAt,
		User:      user.ToResponse(),
	}, nil
}
