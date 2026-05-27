package service

import (
	"context"
	"fmt"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/validator"
	"github.com/sfedu-crm/pkg/hash"
)

type UserService struct {
	userRepo repository.UserRepository
	appRepo  repository.ApplicationRepository
	hasher   *hash.Bcrypt
}

func NewUserService(
	userRepo repository.UserRepository,
	appRepo repository.ApplicationRepository,
	hasher *hash.Bcrypt,
) *UserService {
	return &UserService{userRepo: userRepo, appRepo: appRepo, hasher: hasher}
}

// GetByID — получить пользователя по ID
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// List — список клиентов (для менеджера и админа)
func (s *UserService) List(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	return s.userRepo.List(ctx, filter)
}

// UpdateProfile — клиент редактирует свои данные
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
	return s.userRepo.Update(ctx, id, input)
}

// ChangePassword — смена пароля пользователем
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

// AdminResetPassword — сброс пароля администратором без старого пароля
func (s *UserService) AdminResetPassword(ctx context.Context, targetUserID int64, newPassword string) error {
	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("userService.AdminResetPassword hash: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, targetUserID, newHash)
}

// CreateUser — создание пользователя менеджером/админом
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
	return s.userRepo.Create(ctx, input)
}

// ApproveApplication — принять заявку и создать клиента
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
		Gender:   domain.GenderMale, // по умолчанию, можно обновить позже
	})
	if err != nil {
		return nil, err
	}

	if err := s.appRepo.UpdateStatus(ctx, appID, "approved"); err != nil {
		return nil, err
	}
	return user, nil
}

// SetActive — блокировка/разблокировка пользователя
func (s *UserService) SetActive(ctx context.Context, id int64, active bool) error {
	return s.userRepo.SetActive(ctx, id, active)
}

// Delete — удаление пользователя (только админ)
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}

// ListApplications — список заявок на регистрацию
func (s *UserService) ListApplications(ctx context.Context, status string) ([]*domain.ApplicationRequest, error) {
	return s.appRepo.List(ctx, status)
}

// RejectApplication — отклонить заявку
func (s *UserService) RejectApplication(ctx context.Context, appID int64) error {
	return s.appRepo.UpdateStatus(ctx, appID, "rejected")
}
