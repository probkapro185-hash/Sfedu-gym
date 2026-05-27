package service

import (
	"context"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
)

type TrainerService struct {
	trainerRepo repository.TrainerRepository
}

func NewTrainerService(trainerRepo repository.TrainerRepository) *TrainerService {
	return &TrainerService{trainerRepo: trainerRepo}
}

func (s *TrainerService) List(ctx context.Context, filter repository.TrainerFilter) ([]*domain.Trainer, error) {
	return s.trainerRepo.List(ctx, filter)
}

func (s *TrainerService) GetByID(ctx context.Context, id int64) (*domain.Trainer, error) {
	return s.trainerRepo.GetByID(ctx, id)
}

func (s *TrainerService) Create(ctx context.Context, input domain.CreateTrainerInput) (*domain.Trainer, error) {
	return s.trainerRepo.Create(ctx, input)
}

func (s *TrainerService) Update(ctx context.Context, id int64, input domain.UpdateTrainerInput) (*domain.Trainer, error) {
	return s.trainerRepo.Update(ctx, id, input)
}

func (s *TrainerService) Delete(ctx context.Context, id int64) error {
	return s.trainerRepo.Delete(ctx, id)
}
