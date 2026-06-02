package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/validator"
	"github.com/sfedu-crm/pkg/hash"
	"github.com/sfedu-crm/pkg/jwt"
)

type AuthService struct {
	userRepo repository.UserRepository
	appRepo  repository.ApplicationRepository
	tokenMgr *jwt.Manager
	hasher   *hash.Bcrypt
	tokenTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	appRepo repository.ApplicationRepository,
	tokenMgr *jwt.Manager,
	hasher *hash.Bcrypt,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		appRepo:  appRepo,
		tokenMgr: tokenMgr,
		hasher:   hasher,
		tokenTTL: tokenTTL,
	}
}

// SubmitApplication — публичная форма «Записаться» на главной странице
func (s *AuthService) SubmitApplication(ctx context.Context, input domain.CreateApplicationInput) (*domain.ApplicationRequest, error) {
	if err := validator.ValidateFullName(input.FullName); err != nil {
		return nil, err
	}
	if err := validator.ValidatePhone(input.Phone); err != nil {
		return nil, err
	}
	if err := validator.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	return s.appRepo.Create(ctx, input)
}

// Login — вход по email + пароль, возвращает JWT-токен
func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (string, *domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return "", nil, domain.ErrUnauthorized
	}
	if !user.IsActive {
		return "", nil, fmt.Errorf("%w: account is inactive", domain.ErrForbidden)
	}
	if !s.hasher.Compare(user.PasswordHash, input.Password) {
		return "", nil, domain.ErrInvalidPassword
	}

	token, err := s.tokenMgr.Generate(jwt.Claims{
		UserID: user.ID,
		Role:   string(user.Role),
	}, s.tokenTTL)
	if err != nil {
		return "", nil, fmt.Errorf("authService.Login generate token: %w", err)
	}
	return token, user, nil
}
