package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	const q = `
		INSERT INTO users (full_name, phone, email, password_hash, role, gender)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, full_name, phone, email, role, gender, balance, visits, is_active, created_at, updated_at, last_visit_at`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q,
		input.FullName, input.Phone, input.Email,
		input.Password, input.Role, input.Gender,
	).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email,
		&u.Role, &u.Gender, &u.Balance, &u.Visits,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
	)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Create: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	const q = `
		SELECT id, full_name, phone, email, password_hash, role, gender,
		       balance, visits, is_active, created_at, updated_at, last_visit_at
		FROM users WHERE id = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Gender, &u.Balance, &u.Visits,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, full_name, phone, email, password_hash, role, gender,
		       balance, visits, is_active, created_at, updated_at, last_visit_at
		FROM users WHERE email = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Gender, &u.Balance, &u.Visits,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	const q = `
		SELECT id, full_name, phone, email, password_hash, role, gender,
		       balance, visits, is_active, created_at, updated_at, last_visit_at
		FROM users WHERE phone = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, phone).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Gender, &u.Balance, &u.Visits,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByPhone: %w", err)
	}
	return u, nil
}

func (r *userRepo) List(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *filter.Role)
		argIdx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(full_name ILIKE $%d OR phone ILIKE $%d)", argIdx, argIdx+1,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
		argIdx += 2
	}

	q := `SELECT u.id, u.full_name, u.phone, u.email, u.role, u.gender,
	             u.balance, u.visits, u.is_active, u.created_at, u.updated_at, u.last_visit_at,
	             s.sessions_left
	      FROM users u
	      LEFT JOIN LATERAL (
	        SELECT sessions_left FROM client_subscriptions
	        WHERE client_id = u.id AND is_active = true
	        ORDER BY end_date DESC LIMIT 1
	      ) s ON true`
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY u.created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("userRepo.List: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.FullName, &u.Phone, &u.Email,
			&u.Role, &u.Gender, &u.Balance, &u.Visits,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
			&u.SessionsLeft,
		); err != nil {
			return nil, fmt.Errorf("userRepo.List scan: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepo) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	const q = `
		UPDATE users
		SET full_name=$1, phone=$2, email=$3, gender=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING id, full_name, phone, email, role, gender,
		          balance, visits, is_active, created_at, updated_at, last_visit_at`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, input.FullName, input.Phone, input.Email, input.Gender, id).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email,
		&u.Role, &u.Gender, &u.Balance, &u.Visits,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastVisitAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.Update: %w", err)
	}
	return u, nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	const q = `UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, q, passwordHash, id)
	return err
}

func (r *userRepo) UpdateBalance(ctx context.Context, id int64, delta float64) error {
	const q = `UPDATE users SET balance = balance + $1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, q, delta, id)
	return err
}

func (r *userRepo) IncrementVisits(ctx context.Context, id int64) error {
	const q = `UPDATE users SET visits = visits + 1, last_visit_at=NOW(), updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM users WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *userRepo) SetActive(ctx context.Context, id int64, active bool) error {
	const q = `UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, q, active, id)
	return err
}
