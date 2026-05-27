package domain

import "time"

// SubscriptionType — тип абонемента
type SubscriptionType string

const (
	SubTypeMonthly   SubscriptionType = "monthly"   // Месячный
	SubTypeQuarterly SubscriptionType = "quarterly" // Квартальный
	SubTypeAnnual    SubscriptionType = "annual"    // Годовой
	SubTypeSingle    SubscriptionType = "single"    // Разовое занятие
)

// ProductCategory — категория товара в магазине
type ProductCategory string

const (
	CategorySubscription ProductCategory = "subscription" // Абонементы
	CategorySports       ProductCategory = "sports"       // Спортивные товары
)

// Product — товар в магазине (включая абонементы)
type Product struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	Category    ProductCategory `json:"category"`
	// Поля специфичные для абонементов
	SubType       *SubscriptionType `json:"sub_type,omitempty"`
	DurationDays  *int              `json:"duration_days,omitempty"`
	SessionsCount *int              `json:"sessions_count,omitempty"`
	IsActive      bool              `json:"is_active"`
	CreatedAt     time.Time         `json:"created_at"`
}

// ClientSubscription — купленный абонемент клиента
type ClientSubscription struct {
	ID           int64     `json:"id"`
	ClientID     int64     `json:"client_id"`
	ProductID    int64     `json:"product_id"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	SessionsLeft *int      `json:"sessions_left,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`

	ProductName string  `json:"product_name,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

// Order — заказ в магазине
type Order struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"client_id"`
	ProductID  int64     `json:"product_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"` // pending, paid, cancelled
	CreatedAt  time.Time `json:"created_at"`

	ClientName  string `json:"client_name,omitempty"`
	ProductName string `json:"product_name,omitempty"`
}

// DTO
type CreateProductInput struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Price         float64         `json:"price"`
	Category      ProductCategory `json:"category"`
	SubType       *SubscriptionType `json:"sub_type,omitempty"`
	DurationDays  *int              `json:"duration_days,omitempty"`
	SessionsCount *int              `json:"sessions_count,omitempty"`
}

type PurchaseProductInput struct {
	ProductID int64 `json:"product_id"`
}
