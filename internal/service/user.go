package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/validator"
	pkgcache "github.com/sfedu-crm/pkg/cache"
	"github.com/sfedu-crm/pkg/hash"
)

const (
	usersCacheKey = "users:list"
	usersTTL      = 2 * time.Minute
)

type UserService struct {
	userRepo repository.UserRepository
	appRepo  repository.ApplicationRepository
	hasher   *hash.Bcrypt
	cache    *pkgcache.RedisCache
}

func NewUserService(
	userRepo repository.UserRepository,
	appRepo repository.ApplicationRepository,
	hasher *hash.Bcrypt,
	cache *pkgcache.RedisCache,
) *UserService {
	return &UserService{userRepo: userRepo, appRepo: appRepo, hasher: hasher, cache: cache}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	if s.cache != nil && filter.Role == nil && filter.Search == "" {
		var cached []*domain.User
		if err := s.cache.Get(ctx, usersCacheKey, &cached); err == nil {
			return cached, nil
		}
	}
	users, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && filter.Role == nil && filter.Search == "" {
		_ = s.cache.Set(ctx, usersCacheKey, users, usersTTL)
	}
	return users, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	if err := validator.ValidateFullName(input.FullName); err != nil {
		return nil, err
	}
	if err := validator.ValidatePhone(input.Phone); err != nil {
		return nil, err
	}
	if err := validator.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	user, err := s.userRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, usersCacheKey)
	}
	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id int64, input domain.ChangePasswordInput) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.hasher.Compare(user.PasswordHash, input.OldPassword) {
		return domain.ErrInvalidPassword
	}
	newHash, err := s.hasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("userService.ChangePassword hash: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, id, newHash)
}

func (s *UserService) AdminResetPassword(ctx context.Context, targetUserID int64, newPassword string) error {
	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("userService.AdminResetPassword hash: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, targetUserID, newHash)
}

func (s *UserService) CreateUser(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	if err := validator.ValidateFullName(input.FullName); err != nil {
		return nil, err
	}
	if err := validator.ValidatePhone(input.Phone); err != nil {
		return nil, err
	}
	if err := validator.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("userService.CreateUser hash: %w", err)
	}
	input.Password = passwordHash
	user, err := s.userRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, usersCacheKey)
	}
	return user, nil
}

func (s *UserService) ApproveApplication(ctx context.Context, appID int64, password string) (*domain.User, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app.Status != "pending" {
		return nil, fmt.Errorf("%w: application already processed", domain.ErrInvalidInput)
	}
	user, err := s.CreateUser(ctx, domain.CreateUserInput{
		FullName: app.FullName,
		Phone:    app.Phone,
		Email:    app.Email,
		Password: password,
		Role:     domain.RoleClient,
		Gender:   domain.GenderMale,
	})
	if err != nil {
		return nil, err
	}
	if err := s.appRepo.UpdateStatus(ctx, appID, "approved"); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) SetActive(ctx context.Context, id int64, active bool) error {
	if err := s.userRepo.SetActive(ctx, id, active); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, usersCacheKey)
	}
	return nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, usersCacheKey)
	}
	return nil
}

func (s *UserService) ListApplications(ctx context.Context, status string) ([]*domain.ApplicationRequest, error) {
	return s.appRepo.List(ctx, status)
}

func (s *UserService) RejectApplication(ctx context.Context, appID int64) error {
	return s.appRepo.UpdateStatus(ctx, appID, "rejected")
}
