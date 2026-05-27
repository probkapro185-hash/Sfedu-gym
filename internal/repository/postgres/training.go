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

type trainingRepo struct {
	db *pgxpool.Pool
}

func NewTrainingRepository(db *pgxpool.Pool) repository.TrainingRepository {
	return &trainingRepo{db: db}
}

func (r *trainingRepo) Create(ctx context.Context, input domain.CreateTrainingInput) (*domain.Training, error) {
	const q = `
		INSERT INTO trainings (client_id, trainer_id, title, description, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'scheduled')
		RETURNING id, client_id, trainer_id, title, description, start_time, end_time, status, created_at, updated_at`

	t := &domain.Training{}
	err := r.db.QueryRow(ctx, q,
		input.ClientID, input.TrainerID, input.Title,
		input.Description, input.StartTime, input.EndTime,
	).Scan(
		&t.ID, &t.ClientID, &t.TrainerID, &t.Title,
		&t.Description, &t.StartTime, &t.EndTime,
		&t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("trainingRepo.Create: %w", err)
	}
	return t, nil
}

func (r *trainingRepo) GetByID(ctx context.Context, id int64) (*domain.Training, error) {
	const q = `
		SELECT t.id, t.client_id, t.trainer_id, t.title, t.description,
		       t.start_time, t.end_time, t.status, t.created_at, t.updated_at,
		       u.full_name AS client_name,
		       COALESCE(tr.full_name, '') AS trainer_name
		FROM trainings t
		JOIN users u ON u.id = t.client_id
		LEFT JOIN users tr ON tr.id = t.trainer_id
		WHERE t.id = $1`

	t := &domain.Training{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.ClientID, &t.TrainerID, &t.Title,
		&t.Description, &t.StartTime, &t.EndTime,
		&t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.ClientName, &t.TrainerName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("trainingRepo.GetByID: %w", err)
	}
	return t, nil
}

func (r *trainingRepo) List(ctx context.Context, filter domain.ScheduleFilter) ([]*domain.Training, error) {
	q := `
		SELECT t.id, t.client_id, t.trainer_id, t.title, t.description,
		       t.start_time, t.end_time, t.status, t.created_at, t.updated_at,
		       u.full_name AS client_name,
		       COALESCE(tr.full_name, '') AS trainer_name
		FROM trainings t
		JOIN users u ON u.id = t.client_id
		LEFT JOIN users tr ON tr.id = t.trainer_id
		WHERE t.start_time >= $1 AND t.start_time < $2`

	args := []interface{}{filter.DateFrom, filter.DateTo}
	argIdx := 3

	if filter.ClientID != nil {
		q += fmt.Sprintf(" AND t.client_id = $%d", argIdx)
		args = append(args, *filter.ClientID)
		argIdx++
	}
	if filter.TrainerID != nil {
		q += fmt.Sprintf(" AND t.trainer_id = $%d", argIdx)
		args = append(args, *filter.TrainerID)
		argIdx++
	}
	if filter.Status != "" {
		q += fmt.Sprintf(" AND t.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	q += " ORDER BY t.start_time ASC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("trainingRepo.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Training
	for rows.Next() {
		t := &domain.Training{}
		if err := rows.Scan(
			&t.ID, &t.ClientID, &t.TrainerID, &t.Title,
			&t.Description, &t.StartTime, &t.EndTime,
			&t.Status, &t.CreatedAt, &t.UpdatedAt,
			&t.ClientName, &t.TrainerName,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *trainingRepo) Update(ctx context.Context, id int64, input domain.UpdateTrainingInput) (*domain.Training, error) {
	const q = `
		UPDATE trainings
		SET trainer_id=$1, title=$2, description=$3,
		    start_time=$4, end_time=$5, status=$6, updated_at=NOW()
		WHERE id=$7
		RETURNING id, client_id, trainer_id, title, description,
		          start_time, end_time, status, created_at, updated_at`

	t := &domain.Training{}
	err := r.db.QueryRow(ctx, q,
		input.TrainerID, input.Title, input.Description,
		input.StartTime, input.EndTime, input.Status, id,
	).Scan(
		&t.ID, &t.ClientID, &t.TrainerID, &t.Title,
		&t.Description, &t.StartTime, &t.EndTime,
		&t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("trainingRepo.Update: %w", err)
	}
	return t, nil
}

func (r *trainingRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM trainings WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

// --- Training Requests ---

type trainingRequestRepo struct {
	db *pgxpool.Pool
}

func NewTrainingRequestRepository(db *pgxpool.Pool) repository.TrainingRequestRepository {
	return &trainingRequestRepo{db: db}
}

func (r *trainingRequestRepo) Create(ctx context.Context, clientID int64, input domain.CreateTrainingRequestInput) (*domain.TrainingRequest, error) {
	const q = `
		INSERT INTO training_requests (client_id, preferred_at, comment, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, client_id, preferred_at, comment, status, created_at`

	tr := &domain.TrainingRequest{}
	err := r.db.QueryRow(ctx, q, clientID, input.PreferredAt, input.Comment).Scan(
		&tr.ID, &tr.ClientID, &tr.PreferredAt, &tr.Comment, &tr.Status, &tr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("trainingRequestRepo.Create: %w", err)
	}
	return tr, nil
}

func (r *trainingRequestRepo) GetByID(ctx context.Context, id int64) (*domain.TrainingRequest, error) {
	const q = `
		SELECT tr.id, tr.client_id, tr.preferred_at, tr.comment, tr.status, tr.created_at,
		       u.full_name AS client_name
		FROM training_requests tr
		JOIN users u ON u.id = tr.client_id
		WHERE tr.id=$1`

	t := &domain.TrainingRequest{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.ClientID, &t.PreferredAt, &t.Comment,
		&t.Status, &t.CreatedAt, &t.ClientName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *trainingRequestRepo) ListByClient(ctx context.Context, clientID int64) ([]*domain.TrainingRequest, error) {
	const q = `
		SELECT id, client_id, preferred_at, comment, status, created_at
		FROM training_requests WHERE client_id=$1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.TrainingRequest
	for rows.Next() {
		t := &domain.TrainingRequest{}
		if err := rows.Scan(&t.ID, &t.ClientID, &t.PreferredAt, &t.Comment, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *trainingRequestRepo) ListPending(ctx context.Context) ([]*domain.TrainingRequest, error) {
	const q = `
		SELECT tr.id, tr.client_id, tr.preferred_at, tr.comment, tr.status, tr.created_at,
		       u.full_name AS client_name
		FROM training_requests tr
		JOIN users u ON u.id = tr.client_id
		WHERE tr.status='pending' ORDER BY tr.created_at ASC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.TrainingRequest
	for rows.Next() {
		t := &domain.TrainingRequest{}
		if err := rows.Scan(
			&t.ID, &t.ClientID, &t.PreferredAt, &t.Comment,
			&t.Status, &t.CreatedAt, &t.ClientName,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *trainingRequestRepo) UpdateStatus(ctx context.Context, id int64, status domain.TrainingStatus) error {
	const q = `UPDATE training_requests SET status=$1 WHERE id=$2`
	_, err := r.db.Exec(ctx, q, status, id)
	return err
}
