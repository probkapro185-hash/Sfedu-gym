package domain

import "time"

// OperationType — тип финансовой операции
type OperationType string

const (
	OperationIncome   OperationType = "income"   // Доход (пополнение)
	OperationExpense  OperationType = "expense"  // Расход
	OperationRefund   OperationType = "refund"   // Возврат
)

// ServiceType — тип оплаченной услуги
type ServiceType string

const (
	ServiceSubscription ServiceType = "subscription" // Абонемент
	ServiceTraining     ServiceType = "training"     // Разовая тренировка
	ServiceProduct      ServiceType = "product"      // Товар
	ServiceDeposit      ServiceType = "deposit"      // Пополнение баланса
)

// Payment — запись о финансовой операции
type Payment struct {
	ID            int64         `json:"id"`
	ClientID      int64         `json:"client_id"`
	Amount        float64       `json:"amount"`
	OperationType OperationType `json:"operation_type"`
	ServiceType   ServiceType   `json:"service_type"`
	Description   string        `json:"description"`
	CreatedAt     time.Time     `json:"created_at"`

	// JOIN поля для отображения в финансах (только для админа)
	ClientName string `json:"client_name,omitempty"`
}

// PaymentFilter — фильтр для финансового отчёта
type PaymentFilter struct {
	ClientID      *int64
	OperationType OperationType
	ServiceType   ServiceType
	DateFrom      *time.Time
	DateTo        *time.Time
}

// FinanceSummary — сводка по финансам (только для админа)
type FinanceSummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	TotalRefund  float64 `json:"total_refund"`
	NetBalance   float64 `json:"net_balance"`
}

// DTO
type CreatePaymentInput struct {
	ClientID      int64         `json:"client_id"`
	Amount        float64       `json:"amount"`
	OperationType OperationType `json:"operation_type"`
	ServiceType   ServiceType   `json:"service_type"`
	Description   string        `json:"description"`
}

// TopUpBalanceInput — пополнение баланса клиента
type TopUpBalanceInput struct {
	Amount float64 `json:"amount"`
}
