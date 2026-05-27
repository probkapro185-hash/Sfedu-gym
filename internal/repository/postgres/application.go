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

type applicationRepo struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) repository.ApplicationRepository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) Create(ctx context.Context, input domain.CreateApplicationInput) (*domain.ApplicationRequest, error) {
	const q = `
		INSERT INTO application_requests (full_name, phone, email, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, full_name, phone, email, status, created_at`

	a := &domain.ApplicationRequest{}
	err := r.db.QueryRow(ctx, q, input.FullName, input.Phone, input.Email).Scan(
		&a.ID, &a.FullName, &a.Phone, &a.Email, &a.Status, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("applicationRepo.Create: %w", err)
	}
	return a, nil
}

func (r *applicationRepo) GetByID(ctx context.Context, id int64) (*domain.ApplicationRequest, error) {
	const q = `SELECT id, full_name, phone, email, status, created_at FROM application_requests WHERE id=$1`
	a := &domain.ApplicationRequest{}
	err := r.db.QueryRow(ctx, q, id).Scan(&a.ID, &a.FullName, &a.Phone, &a.Email, &a.Status, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("applicationRepo.GetByID: %w", err)
	}
	return a, nil
}

func (r *applicationRepo) List(ctx context.Context, status string) ([]*domain.ApplicationRequest, error) {
	q := `SELECT id, full_name, phone, email, status, created_at FROM application_requests`
	var args []interface{}
	if status != "" {
		q += " WHERE status=$1"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("applicationRepo.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.ApplicationRequest
	for rows.Next() {
		a := &domain.ApplicationRequest{}
		if err := rows.Scan(&a.ID, &a.FullName, &a.Phone, &a.Email, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *applicationRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const q = `UPDATE application_requests SET status=$1 WHERE id=$2`
	_, err := r.db.Exec(ctx, q, status, id)
	return err
}

func (r *applicationRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM application_requests WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
