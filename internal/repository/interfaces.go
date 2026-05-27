package repository

import (
	"context"

	"github.com/sfedu-crm/internal/domain"
)

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	List(ctx context.Context, filter UserFilter) ([]*domain.User, error)
	Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateBalance(ctx context.Context, id int64, delta float64) error
	IncrementVisits(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	SetActive(ctx context.Context, id int64, active bool) error
}

type UserFilter struct {
	Role     *domain.Role
	IsActive *bool
	Search   string // поиск по ФИО или телефону
}

// ApplicationRepository — заявки на регистрацию от клиентов
type ApplicationRepository interface {
	Create(ctx context.Context, input domain.CreateApplicationInput) (*domain.ApplicationRequest, error)
	GetByID(ctx context.Context, id int64) (*domain.ApplicationRequest, error)
	List(ctx context.Context, status string) ([]*domain.ApplicationRequest, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
}

// TrainerRepository — профили тренеров
type TrainerRepository interface {
	Create(ctx context.Context, input domain.CreateTrainerInput) (*domain.Trainer, error)
	GetByID(ctx context.Context, id int64) (*domain.Trainer, error)
	GetByUserID(ctx context.Context, userID int64) (*domain.Trainer, error)
	List(ctx context.Context, filter TrainerFilter) ([]*domain.Trainer, error)
	Update(ctx context.Context, id int64, input domain.UpdateTrainerInput) (*domain.Trainer, error)
	Delete(ctx context.Context, id int64) error
}

type TrainerFilter struct {
	Specialization *domain.TrainerSpecialization
	IsActive       *bool
	Search         string
}

// TrainingRepository — занятия / расписание
type TrainingRepository interface {
	Create(ctx context.Context, input domain.CreateTrainingInput) (*domain.Training, error)
	GetByID(ctx context.Context, id int64) (*domain.Training, error)
	List(ctx context.Context, filter domain.ScheduleFilter) ([]*domain.Training, error)
	Update(ctx context.Context, id int64, input domain.UpdateTrainingInput) (*domain.Training, error)
	Delete(ctx context.Context, id int64) error
}

// TrainingRequestRepository — заявки клиентов на тренировки
type TrainingRequestRepository interface {
	Create(ctx context.Context, clientID int64, input domain.CreateTrainingRequestInput) (*domain.TrainingRequest, error)
	GetByID(ctx context.Context, id int64) (*domain.TrainingRequest, error)
	ListByClient(ctx context.Context, clientID int64) ([]*domain.TrainingRequest, error)
	ListPending(ctx context.Context) ([]*domain.TrainingRequest, error)
	UpdateStatus(ctx context.Context, id int64, status domain.TrainingStatus) error
}

// ProductRepository — магазин: абонементы и товары
type ProductRepository interface {
	Create(ctx context.Context, input domain.CreateProductInput) (*domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, filter ProductFilter) ([]*domain.Product, error)
	Update(ctx context.Context, id int64, input domain.CreateProductInput) (*domain.Product, error)
	Delete(ctx context.Context, id int64) error
}

type ProductFilter struct {
	Category *domain.ProductCategory
	IsActive *bool
}

// SubscriptionRepository — купленные абонементы клиентов
type SubscriptionRepository interface {
	Create(ctx context.Context, clientID, productID int64) (*domain.ClientSubscription, error)
	GetByID(ctx context.Context, id int64) (*domain.ClientSubscription, error)
	ListByClient(ctx context.Context, clientID int64) ([]*domain.ClientSubscription, error)
	GetActiveByClient(ctx context.Context, clientID int64) (*domain.ClientSubscription, error)
	DecrementSessions(ctx context.Context, id int64) error
	Deactivate(ctx context.Context, id int64) error
}

// PaymentRepository — финансовые операции
type PaymentRepository interface {
	Create(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error)
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	List(ctx context.Context, filter domain.PaymentFilter) ([]*domain.Payment, error)
	GetSummary(ctx context.Context, filter domain.PaymentFilter) (*domain.FinanceSummary, error)
}
