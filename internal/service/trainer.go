package service

import (
	"context"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	pkgcache "github.com/sfedu-crm/pkg/cache"
)

const (
	trainersCacheKey = "trainers:list"
	trainersTTL      = 5 * time.Minute
)

type TrainerService struct {
	trainerRepo repository.TrainerRepository
	cache       *pkgcache.RedisCache
}

func NewTrainerService(trainerRepo repository.TrainerRepository, cache *pkgcache.RedisCache) *TrainerService {
	return &TrainerService{trainerRepo: trainerRepo, cache: cache}
}

func (s *TrainerService) List(ctx context.Context, filter repository.TrainerFilter) ([]*domain.Trainer, error) {
	// Пробуем получить из кэша только если нет фильтров
	if s.cache != nil && filter.Specialization == nil && filter.Search == "" {
		var cached []*domain.Trainer
		if err := s.cache.Get(ctx, trainersCacheKey, &cached); err == nil {
			return cached, nil
		}
	}

	trainers, err := s.trainerRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	if s.cache != nil && filter.Specialization == nil && filter.Search == "" {
		_ = s.cache.Set(ctx, trainersCacheKey, trainers, trainersTTL)
	}

	return trainers, nil
}

func (s *TrainerService) GetByID(ctx context.Context, id int64) (*domain.Trainer, error) {
	return s.trainerRepo.GetByID(ctx, id)
}

func (s *TrainerService) Create(ctx context.Context, input domain.CreateTrainerInput) (*domain.Trainer, error) {
	trainer, err := s.trainerRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	// Инвалидируем кэш
	if s.cache != nil {
		_ = s.cache.Delete(ctx, trainersCacheKey)
	}
	return trainer, nil
}

func (s *TrainerService) Update(ctx context.Context, id int64, input domain.UpdateTrainerInput) (*domain.Trainer, error) {
	trainer, err := s.trainerRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, trainersCacheKey)
	}
	return trainer, nil
}

func (s *TrainerService) Delete(ctx context.Context, id int64) error {
	if err := s.trainerRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, trainersCacheKey)
	}
	return nil
}
