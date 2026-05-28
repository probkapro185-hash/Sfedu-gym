package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/service"
	pkghash "github.com/sfedu-crm/pkg/hash"
	"github.com/sfedu-crm/pkg/jwt"
)

// ── MockUserRepo ──────────────────────────────────────────────────────────────
type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) List(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.User), args.Error(1)
}
func (m *MockUserRepo) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) UpdatePassword(ctx context.Context, id int64, h string) error {
	return m.Called(ctx, id, h).Error(0)
}
func (m *MockUserRepo) UpdateBalance(ctx context.Context, id int64, delta float64) error {
	return m.Called(ctx, id, delta).Error(0)
}
func (m *MockUserRepo) IncrementVisits(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepo) SetActive(ctx context.Context, id int64, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}

// ── MockAppRepo ───────────────────────────────────────────────────────────────
type MockAppRepo struct{ mock.Mock }

func (m *MockAppRepo) Create(ctx context.Context, input domain.CreateApplicationInput) (*domain.ApplicationRequest, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ApplicationRequest), args.Error(1)
}
func (m *MockAppRepo) GetByID(ctx context.Context, id int64) (*domain.ApplicationRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ApplicationRequest), args.Error(1)
}
func (m *MockAppRepo) List(ctx context.Context, status string) ([]*domain.ApplicationRequest, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.ApplicationRequest), args.Error(1)
}
func (m *MockAppRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *MockAppRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

// ── helpers ───────────────────────────────────────────────────────────────────
func newTestAuthService(userRepo repository.UserRepository, appRepo repository.ApplicationRepository) *service.AuthService {
	tokenMgr := jwt.NewManager("test-secret")
	hasher := pkghash.NewBcrypt(12)
	return service.NewAuthService(userRepo, appRepo, tokenMgr, hasher, 24*time.Hour)
}

// ── tests ─────────────────────────────────────────────────────────────────────
func TestLogin_Success(t *testing.T) {
	password := "Admin1234"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	user := &domain.User{
		ID: 1, Email: "admin@gmail.com",
		PasswordHash: string(hashed),
		Role:         domain.RoleAdmin, IsActive: true,
	}
	mockUser := new(MockUserRepo)
	mockApp := new(MockAppRepo)
	mockUser.On("GetByEmail", mock.Anything, "admin@gmail.com").Return(user, nil)

	svc := newTestAuthService(mockUser, mockApp)
	token, result, err := svc.Login(context.Background(), domain.LoginInput{
		Email: "admin@gmail.com", Password: password,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, user.ID, result.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), 12)
	user := &domain.User{
		ID: 1, Email: "test@gmail.com",
		PasswordHash: string(hashed), IsActive: true,
	}
	mockUser := new(MockUserRepo)
	mockApp := new(MockAppRepo)
	mockUser.On("GetByEmail", mock.Anything, "test@gmail.com").Return(user, nil)

	svc := newTestAuthService(mockUser, mockApp)
	token, result, err := svc.Login(context.Background(), domain.LoginInput{
		Email: "test@gmail.com", Password: "wrong",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, result)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUser := new(MockUserRepo)
	mockApp := new(MockAppRepo)
	mockUser.On("GetByEmail", mock.Anything, "notfound@gmail.com").Return(nil, domain.ErrNotFound)

	svc := newTestAuthService(mockUser, mockApp)
	token, result, err := svc.Login(context.Background(), domain.LoginInput{
		Email: "notfound@gmail.com", Password: "pass",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, result)
}

func TestLogin_InactiveUser(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), 12)
	user := &domain.User{
		ID: 1, Email: "blocked@gmail.com",
		PasswordHash: string(hashed), IsActive: false,
	}
	mockUser := new(MockUserRepo)
	mockApp := new(MockAppRepo)
	mockUser.On("GetByEmail", mock.Anything, "blocked@gmail.com").Return(user, nil)

	svc := newTestAuthService(mockUser, mockApp)
	token, result, err := svc.Login(context.Background(), domain.LoginInput{
		Email: "blocked@gmail.com", Password: "pass",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, result)
}
