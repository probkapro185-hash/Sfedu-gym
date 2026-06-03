package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
)

type ScheduleService struct {
	trainingRepo     repository.TrainingRepository
	requestRepo      repository.TrainingRequestRepository
	userRepo         repository.UserRepository
	paymentRepo      repository.PaymentRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewScheduleService(
	trainingRepo repository.TrainingRepository,
	requestRepo repository.TrainingRequestRepository,
	userRepo repository.UserRepository,
	paymentRepo repository.PaymentRepository,
	subscriptionRepo repository.SubscriptionRepository,
) *ScheduleService {
	return &ScheduleService{
		trainingRepo:     trainingRepo,
		requestRepo:      requestRepo,
		userRepo:         userRepo,
		paymentRepo:      paymentRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// SubmitTrainingRequest — клиент подаёт заявку на тренировку
func (s *ScheduleService) SubmitTrainingRequest(ctx context.Context, clientID int64, input domain.CreateTrainingRequestInput) (*domain.TrainingRequest, error) {
	if input.PreferredAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w: preferred time must be in the future", domain.ErrInvalidInput)
	}
	return s.requestRepo.Create(ctx, clientID, input)
}

// ListPendingRequests — список заявок (для менеджера/админа)
func (s *ScheduleService) ListPendingRequests(ctx context.Context) ([]*domain.TrainingRequest, error) {
	return s.requestRepo.ListPending(ctx)
}

// ListClientRequests — заявки конкретного клиента
func (s *ScheduleService) ListClientRequests(ctx context.Context, clientID int64) ([]*domain.TrainingRequest, error) {
	return s.requestRepo.ListByClient(ctx, clientID)
}

// ApproveRequest — менеджер/админ принимает заявку и создаёт занятие
func (s *ScheduleService) ApproveRequest(ctx context.Context, requestID int64, input domain.CreateTrainingInput) (*domain.Training, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != domain.TrainingStatusPending {
		return nil, fmt.Errorf("%w: request already processed", domain.ErrInvalidInput)
	}

	training, err := s.trainingRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := s.requestRepo.UpdateStatus(ctx, requestID, domain.TrainingStatusScheduled); err != nil {
		return nil, err
	}
	return training, nil
}

// RejectRequest — отклонить заявку
func (s *ScheduleService) RejectRequest(ctx context.Context, requestID int64) error {
	return s.requestRepo.UpdateStatus(ctx, requestID, domain.TrainingStatusCancelled)
}

// GetSchedule — получить расписание (по месяцу/неделе/дню через фильтр)
func (s *ScheduleService) GetSchedule(ctx context.Context, filter domain.ScheduleFilter) ([]*domain.Training, error) {
	return s.trainingRepo.List(ctx, filter)
}

func (s *ScheduleService) CreateTraining(ctx context.Context, input domain.CreateTrainingInput) (*domain.Training, error) {
	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("%w: end time must be after start time", domain.ErrInvalidInput)
	}
	return s.trainingRepo.Create(ctx, input)
}

// GetTrainingByID — детали занятия
func (s *ScheduleService) GetTrainingByID(ctx context.Context, id int64) (*domain.Training, error) {
	return s.trainingRepo.GetByID(ctx, id)
}

// UpdateTraining — перенести занятие (дата/время/тренер) — менеджер/админ
func (s *ScheduleService) UpdateTraining(ctx context.Context, trainingID int64, input domain.UpdateTrainingInput) (*domain.Training, error) {
	if !input.StartTime.IsZero() && !input.EndTime.IsZero() {
		if input.EndTime.Before(input.StartTime) {
			return nil, fmt.Errorf("%w: end time must be after start time", domain.ErrInvalidInput)
		}
	}

	// Получаем текущее занятие чтобы проверить смену статуса
	current, err := s.trainingRepo.GetByID(ctx, trainingID)
	if err != nil {
		return nil, err
	}

	training, err := s.trainingRepo.Update(ctx, trainingID, input)
	if err != nil {
		return nil, err
	}

	// Если статус изменился на "completed" — списываем занятие с абонемента
	if current.Status != domain.TrainingStatusCompleted && input.Status == domain.TrainingStatusCompleted {
		sub, err := s.subscriptionRepo.GetActiveByClient(ctx, training.ClientID)
		if err == nil && sub != nil {
			_ = s.subscriptionRepo.DecrementSessions(ctx, sub.ID)
			_ = s.userRepo.IncrementVisits(ctx, training.ClientID)
		}
	}

	return training, nil
}

// CancelTraining — отменить занятие
func (s *ScheduleService) CancelTraining(ctx context.Context, trainingID int64) error {
	t, err := s.trainingRepo.GetByID(ctx, trainingID)
	if err != nil {
		return err
	}
	cancelInput := domain.UpdateTrainingInput{
		TrainerID:   t.TrainerID,
		Title:       t.Title,
		Description: t.Description,
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
		Status:      domain.TrainingStatusCancelled,
	}
	_, err = s.trainingRepo.Update(ctx, trainingID, cancelInput)
	return err
}

// DeleteTraining — удалить занятие (только админ)
func (s *ScheduleService) DeleteTraining(ctx context.Context, trainingID int64) error {
	return s.trainingRepo.Delete(ctx, trainingID)
}
