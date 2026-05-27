package service

import (
	"context"
	"fmt"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
)

// FinanceService — работа с финансами (только для админа)
type FinanceService struct {
	paymentRepo repository.PaymentRepository
	userRepo    repository.UserRepository
}

func NewFinanceService(paymentRepo repository.PaymentRepository, userRepo repository.UserRepository) *FinanceService {
	return &FinanceService{paymentRepo: paymentRepo, userRepo: userRepo}
}

// TopUpBalance — пополнение баланса клиента
func (s *FinanceService) TopUpBalance(ctx context.Context, clientID int64, input domain.TopUpBalanceInput) (*domain.Payment, error) {
	if input.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}
	if err := s.userRepo.UpdateBalance(ctx, clientID, input.Amount); err != nil {
		return nil, err
	}
	return s.paymentRepo.Create(ctx, domain.CreatePaymentInput{
		ClientID:      clientID,
		Amount:        input.Amount,
		OperationType: domain.OperationIncome,
		ServiceType:   domain.ServiceDeposit,
		Description:   "Пополнение баланса",
	})
}

// ListPayments — история операций (фильтрация для админа)
func (s *FinanceService) ListPayments(ctx context.Context, filter domain.PaymentFilter) ([]*domain.Payment, error) {
	return s.paymentRepo.List(ctx, filter)
}

// GetSummary — сводка по финансам
func (s *FinanceService) GetSummary(ctx context.Context, filter domain.PaymentFilter) (*domain.FinanceSummary, error) {
	return s.paymentRepo.GetSummary(ctx, filter)
}

// GetClientPayments — история платежей конкретного клиента (видит сам клиент)
func (s *FinanceService) GetClientPayments(ctx context.Context, clientID int64) ([]*domain.Payment, error) {
	return s.paymentRepo.List(ctx, domain.PaymentFilter{ClientID: &clientID})
}

// ShopService — магазин (абонементы и товары)
type ShopService struct {
	productRepo      repository.ProductRepository
	subscriptionRepo repository.SubscriptionRepository
	paymentRepo      repository.PaymentRepository
	userRepo         repository.UserRepository
}

func NewShopService(
	productRepo repository.ProductRepository,
	subscriptionRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
) *ShopService {
	return &ShopService{
		productRepo:      productRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		userRepo:         userRepo,
	}
}

// ListProducts — каталог товаров/абонементов
func (s *ShopService) ListProducts(ctx context.Context, filter repository.ProductFilter) ([]*domain.Product, error) {
	return s.productRepo.List(ctx, filter)
}

// GetProduct — детали товара
func (s *ShopService) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

// CreateProduct — добавить товар/абонемент (менеджер/админ)
func (s *ShopService) CreateProduct(ctx context.Context, input domain.CreateProductInput) (*domain.Product, error) {
	if input.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive", domain.ErrInvalidInput)
	}
	return s.productRepo.Create(ctx, input)
}

// UpdateProduct — обновить товар
func (s *ShopService) UpdateProduct(ctx context.Context, id int64, input domain.CreateProductInput) (*domain.Product, error) {
	return s.productRepo.Update(ctx, id, input)
}

// DeleteProduct — удалить товар (только админ)
func (s *ShopService) DeleteProduct(ctx context.Context, id int64) error {
	return s.productRepo.Delete(ctx, id)
}

// PurchaseProduct — клиент покупает абонемент/товар со своего баланса
func (s *ShopService) PurchaseProduct(ctx context.Context, clientID int64, input domain.PurchaseProductInput) (*domain.ClientSubscription, error) {
	product, err := s.productRepo.GetByID(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if user.Balance < product.Price {
		return nil, domain.ErrInsufficientFunds
	}

	// Снимаем деньги
	if err := s.userRepo.UpdateBalance(ctx, clientID, -product.Price); err != nil {
		return nil, err
	}

	// Записываем в историю платежей
	_, err = s.paymentRepo.Create(ctx, domain.CreatePaymentInput{
		ClientID:      clientID,
		Amount:        product.Price,
		OperationType: domain.OperationExpense,
		ServiceType:   domain.ServiceSubscription,
		Description:   fmt.Sprintf("Покупка: %s", product.Name),
	})
	if err != nil {
		return nil, err
	}

	// Создаём абонемент (только для категории subscription)
	if product.Category == domain.CategorySubscription {
		return s.subscriptionRepo.Create(ctx, clientID, input.ProductID)
	}
	return nil, nil
}

// GetClientSubscriptions — список абонементов клиента
func (s *ShopService) GetClientSubscriptions(ctx context.Context, clientID int64) ([]*domain.ClientSubscription, error) {
	return s.subscriptionRepo.ListByClient(ctx, clientID)
}
