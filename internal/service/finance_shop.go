package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	pkgcache "github.com/sfedu-crm/pkg/cache"
)

// ── FinanceService ────────────────────────────────────────────────────────────

type FinanceService struct {
	paymentRepo repository.PaymentRepository
	userRepo    repository.UserRepository
}

func NewFinanceService(paymentRepo repository.PaymentRepository, userRepo repository.UserRepository) *FinanceService {
	return &FinanceService{paymentRepo: paymentRepo, userRepo: userRepo}
}

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

func (s *FinanceService) ListPayments(ctx context.Context, filter domain.PaymentFilter) ([]*domain.Payment, error) {
	return s.paymentRepo.List(ctx, filter)
}

func (s *FinanceService) GetSummary(ctx context.Context, filter domain.PaymentFilter) (*domain.FinanceSummary, error) {
	return s.paymentRepo.GetSummary(ctx, filter)
}

func (s *FinanceService) GetClientPayments(ctx context.Context, clientID int64) ([]*domain.Payment, error) {
	return s.paymentRepo.List(ctx, domain.PaymentFilter{ClientID: &clientID})
}

// ── ShopService ───────────────────────────────────────────────────────────────

const (
	productsCacheKey = "products:list"
	productsTTL      = 10 * time.Minute
)

type ShopService struct {
	productRepo      repository.ProductRepository
	subscriptionRepo repository.SubscriptionRepository
	paymentRepo      repository.PaymentRepository
	userRepo         repository.UserRepository
	cache            *pkgcache.RedisCache
}

func NewShopService(
	productRepo repository.ProductRepository,
	subscriptionRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
	cache *pkgcache.RedisCache,
) *ShopService {
	return &ShopService{
		productRepo:      productRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		userRepo:         userRepo,
		cache:            cache,
	}
}

func (s *ShopService) ListProducts(ctx context.Context, filter repository.ProductFilter) ([]*domain.Product, error) {
	if s.cache != nil && filter.Category == nil && filter.IsActive == nil {
		var cached []*domain.Product
		if err := s.cache.Get(ctx, productsCacheKey, &cached); err == nil {
			return cached, nil
		}
	}
	products, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && filter.Category == nil && filter.IsActive == nil {
		_ = s.cache.Set(ctx, productsCacheKey, products, productsTTL)
	}
	return products, nil
}

func (s *ShopService) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ShopService) CreateProduct(ctx context.Context, input domain.CreateProductInput) (*domain.Product, error) {
	if input.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive", domain.ErrInvalidInput)
	}
	product, err := s.productRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, productsCacheKey)
	}
	return product, nil
}

func (s *ShopService) UpdateProduct(ctx context.Context, id int64, input domain.CreateProductInput) (*domain.Product, error) {
	product, err := s.productRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, productsCacheKey)
	}
	return product, nil
}

func (s *ShopService) DeleteProduct(ctx context.Context, id int64) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, productsCacheKey)
	}
	return nil
}

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
	if err := s.userRepo.UpdateBalance(ctx, clientID, -product.Price); err != nil {
		return nil, err
	}
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
	if product.Category == domain.CategorySubscription {
		return s.subscriptionRepo.Create(ctx, clientID, input.ProductID)
	}
	return nil, nil
}

func (s *ShopService) GetClientSubscriptions(ctx context.Context, clientID int64) ([]*domain.ClientSubscription, error) {
	return s.subscriptionRepo.ListByClient(ctx, clientID)
}
