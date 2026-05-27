package domain

import "time"

// TrainingStatus — статус занятия
type TrainingStatus string

const (
	TrainingStatusScheduled  TrainingStatus = "scheduled"  // Запланировано
	TrainingStatusCompleted  TrainingStatus = "completed"  // Завершено
	TrainingStatusCancelled  TrainingStatus = "cancelled"  // Отменено
	TrainingStatusPending    TrainingStatus = "pending"    // Ожидает подтверждения (заявка)
)

// Training — запись на тренировку / занятие
type Training struct {
	ID          int64          `json:"id"`
	ClientID    int64          `json:"client_id"`
	TrainerID   *int64         `json:"trainer_id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Status      TrainingStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Дополнительные поля для ответов (JOIN-ы)
	ClientName  string `json:"client_name,omitempty"`
	TrainerName string `json:"trainer_name,omitempty"`
}

// TrainingRequest — заявка клиента на тренировку (до принятия менеджером/админом)
type TrainingRequest struct {
	ID          int64          `json:"id"`
	ClientID    int64          `json:"client_id"`
	PreferredAt time.Time      `json:"preferred_at"`
	Comment     string         `json:"comment"`
	Status      TrainingStatus `json:"status"` // pending, scheduled, rejected
	CreatedAt   time.Time      `json:"created_at"`

	ClientName string `json:"client_name,omitempty"`
}

// DTO
type CreateTrainingRequestInput struct {
	PreferredAt time.Time `json:"preferred_at"`
	Comment     string    `json:"comment"`
}

type CreateTrainingInput struct {
	ClientID    int64          `json:"client_id"`
	TrainerID   *int64         `json:"trainer_id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
}

type UpdateTrainingInput struct {
	TrainerID   *int64         `json:"trainer_id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Status      TrainingStatus `json:"status"`
}

// ScheduleFilter — фильтр для получения расписания
type ScheduleFilter struct {
	ClientID  *int64
	TrainerID *int64
	DateFrom  time.Time
	DateTo    time.Time
	Status    TrainingStatus
}
