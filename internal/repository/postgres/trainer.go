package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
)

type trainerRepo struct {
	db *pgxpool.Pool
}

func NewTrainerRepository(db *pgxpool.Pool) repository.TrainerRepository {
	return &trainerRepo{db: db}
}

func (r *trainerRepo) Create(ctx context.Context, input domain.CreateTrainerInput) (*domain.Trainer, error) {
	const q = `
		INSERT INTO trainers (user_id, specialization, bio, photo_url, experience_years)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, user_id, specialization, bio, photo_url, experience_years, is_active, created_at`

	t := &domain.Trainer{}
	// full_name is joined from users
	err := r.db.QueryRow(ctx, q,
		input.UserID, input.Specialization, input.Bio, input.PhotoURL, input.ExperienceYears,
	).Scan(&t.ID, &t.UserID, &t.Specialization, &t.Bio, &t.PhotoURL, &t.ExperienceYears, &t.IsActive, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("trainerRepo.Create: %w", err)
	}
	return t, nil
}

func (r *trainerRepo) GetByID(ctx context.Context, id int64) (*domain.Trainer, error) {
	const q = `
		SELECT t.id, t.user_id, u.full_name, t.specialization, t.bio, t.photo_url, t.experience_years, t.is_active, t.created_at
		FROM trainers t JOIN users u ON u.id=t.user_id WHERE t.id=$1`

	t := &domain.Trainer{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.UserID, &t.FullName, &t.Specialization, &t.Bio, &t.PhotoURL, &t.ExperienceYears, &t.IsActive, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *trainerRepo) GetByUserID(ctx context.Context, userID int64) (*domain.Trainer, error) {
	const q = `
		SELECT t.id, t.user_id, u.full_name, t.specialization, t.bio, t.photo_url, t.experience_years, t.is_active, t.created_at
		FROM trainers t JOIN users u ON u.id=t.user_id WHERE t.user_id=$1`

	t := &domain.Trainer{}
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&t.ID, &t.UserID, &t.FullName, &t.Specialization, &t.Bio, &t.PhotoURL, &t.ExperienceYears, &t.IsActive, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *trainerRepo) List(ctx context.Context, filter repository.TrainerFilter) ([]*domain.Trainer, error) {
	q := `SELECT t.id, t.user_id, u.full_name, t.specialization, t.bio, t.photo_url, t.experience_years, t.is_active, t.created_at
	      FROM trainers t JOIN users u ON u.id=t.user_id WHERE 1=1`

	var args []interface{}
	idx := 1

	if filter.Specialization != nil {
		q += fmt.Sprintf(" AND t.specialization=$%d", idx)
		args = append(args, *filter.Specialization)
		idx++
	}
	if filter.IsActive != nil {
		q += fmt.Sprintf(" AND t.is_active=$%d", idx)
		args = append(args, *filter.IsActive)
		idx++
	}
	if filter.Search != "" {
		q += fmt.Sprintf(" AND u.full_name ILIKE $%d", idx)
		args = append(args, "%"+filter.Search+"%")
		idx++
	}
	q += " ORDER BY t.created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("trainerRepo.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Trainer
	for rows.Next() {
		t := &domain.Trainer{}
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.FullName, &t.Specialization, &t.Bio, &t.PhotoURL, &t.ExperienceYears, &t.IsActive, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *trainerRepo) Update(ctx context.Context, id int64, input domain.UpdateTrainerInput) (*domain.Trainer, error) {
	const q = `
		UPDATE trainers SET specialization=$1, bio=$2, photo_url=$3, experience_years=$4, is_active=$5
		WHERE id=$6
		RETURNING id, user_id, specialization, bio, photo_url, experience_years, is_active, created_at`

	t := &domain.Trainer{}
	err := r.db.QueryRow(ctx, q,
		input.Specialization, input.Bio, input.PhotoURL, input.ExperienceYears, input.IsActive, id,
	).Scan(&t.ID, &t.UserID, &t.Specialization, &t.Bio, &t.PhotoURL, &t.ExperienceYears, &t.IsActive, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *trainerRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM trainers WHERE id=$1`, id)
	return err
}
