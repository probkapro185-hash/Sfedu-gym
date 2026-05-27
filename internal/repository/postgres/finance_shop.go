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

// ---- Payment Repository ----

type paymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) repository.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error) {
	const q = `
		INSERT INTO payments (client_id, amount, operation_type, service_type, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, client_id, amount, operation_type, service_type, description, created_at`

	p := &domain.Payment{}
	err := r.db.QueryRow(ctx, q,
		input.ClientID, input.Amount,
		input.OperationType, input.ServiceType, input.Description,
	).Scan(
		&p.ID, &p.ClientID, &p.Amount,
		&p.OperationType, &p.ServiceType, &p.Description, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.Create: %w", err)
	}
	return p, nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	const q = `
		SELECT p.id, p.client_id, p.amount, p.operation_type, p.service_type, p.description, p.created_at,
		       u.full_name AS client_name
		FROM payments p
		JOIN users u ON u.id = p.client_id
		WHERE p.id=$1`

	p := &domain.Payment{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.ClientID, &p.Amount,
		&p.OperationType, &p.ServiceType, &p.Description, &p.CreatedAt,
		&p.ClientName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func (r *paymentRepo) List(ctx context.Context, filter domain.PaymentFilter) ([]*domain.Payment, error) {
	var conds []string
	var args []interface{}
	idx := 1

	if filter.ClientID != nil {
		conds = append(conds, fmt.Sprintf("p.client_id=$%d", idx))
		args = append(args, *filter.ClientID)
		idx++
	}
	if filter.OperationType != "" {
		conds = append(conds, fmt.Sprintf("p.operation_type=$%d", idx))
		args = append(args, filter.OperationType)
		idx++
	}
	if filter.ServiceType != "" {
		conds = append(conds, fmt.Sprintf("p.service_type=$%d", idx))
		args = append(args, filter.ServiceType)
		idx++
	}
	if filter.DateFrom != nil {
		conds = append(conds, fmt.Sprintf("p.created_at >= $%d", idx))
		args = append(args, *filter.DateFrom)
		idx++
	}
	if filter.DateTo != nil {
		conds = append(conds, fmt.Sprintf("p.created_at < $%d", idx))
		args = append(args, *filter.DateTo)
		idx++
	}

	q := `SELECT p.id, p.client_id, p.amount, p.operation_type, p.service_type,
	             p.description, p.created_at, u.full_name AS client_name
	      FROM payments p
	      JOIN users u ON u.id = p.client_id`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY p.created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(
			&p.ID, &p.ClientID, &p.Amount,
			&p.OperationType, &p.ServiceType, &p.Description, &p.CreatedAt,
			&p.ClientName,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *paymentRepo) GetSummary(ctx context.Context, filter domain.PaymentFilter) (*domain.FinanceSummary, error) {
	var conds []string
	var args []interface{}
	idx := 1

	if filter.ClientID != nil {
		conds = append(conds, fmt.Sprintf("client_id=$%d", idx))
		args = append(args, *filter.ClientID)
		idx++
	}
	if filter.DateFrom != nil {
		conds = append(conds, fmt.Sprintf("created_at>=$%d", idx))
		args = append(args, *filter.DateFrom)
		idx++
	}
	if filter.DateTo != nil {
		conds = append(conds, fmt.Sprintf("created_at<$%d", idx))
		args = append(args, *filter.DateTo)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	q := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE operation_type='income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE operation_type='expense'), 0),
			COALESCE(SUM(amount) FILTER (WHERE operation_type='refund'), 0)
		FROM payments %s`, where)

	s := &domain.FinanceSummary{}
	err := r.db.QueryRow(ctx, q, args...).Scan(&s.TotalIncome, &s.TotalExpense, &s.TotalRefund)
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.GetSummary: %w", err)
	}
	s.NetBalance = s.TotalIncome - s.TotalExpense - s.TotalRefund
	return s, nil
}

// ---- Product Repository ----

type productRepo struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, input domain.CreateProductInput) (*domain.Product, error) {
	const q = `
		INSERT INTO products (name, description, price, category, sub_type, duration_days, sessions_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, name, description, price, category, sub_type, duration_days, sessions_count, is_active, created_at`

	p := &domain.Product{}
	err := r.db.QueryRow(ctx, q,
		input.Name, input.Description, input.Price, input.Category,
		input.SubType, input.DurationDays, input.SessionsCount,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Category,
		&p.SubType, &p.DurationDays, &p.SessionsCount, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("productRepo.Create: %w", err)
	}
	return p, nil
}

func (r *productRepo) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	const q = `SELECT id, name, description, price, category, sub_type, duration_days, sessions_count, is_active, created_at FROM products WHERE id=$1`
	p := &domain.Product{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Category,
		&p.SubType, &p.DurationDays, &p.SessionsCount, &p.IsActive, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func (r *productRepo) List(ctx context.Context, filter repository.ProductFilter) ([]*domain.Product, error) {
	var conds []string
	var args []interface{}
	idx := 1

	if filter.Category != nil {
		conds = append(conds, fmt.Sprintf("category=$%d", idx))
		args = append(args, *filter.Category)
		idx++
	}
	if filter.IsActive != nil {
		conds = append(conds, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *filter.IsActive)
		idx++
	}

	q := `SELECT id, name, description, price, category, sub_type, duration_days, sessions_count, is_active, created_at FROM products`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("productRepo.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Product
	for rows.Next() {
		p := &domain.Product{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Category,
			&p.SubType, &p.DurationDays, &p.SessionsCount, &p.IsActive, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *productRepo) Update(ctx context.Context, id int64, input domain.CreateProductInput) (*domain.Product, error) {
	const q = `
		UPDATE products SET name=$1, description=$2, price=$3, category=$4,
		    sub_type=$5, duration_days=$6, sessions_count=$7
		WHERE id=$8
		RETURNING id, name, description, price, category, sub_type, duration_days, sessions_count, is_active, created_at`

	p := &domain.Product{}
	err := r.db.QueryRow(ctx, q,
		input.Name, input.Description, input.Price, input.Category,
		input.SubType, input.DurationDays, input.SessionsCount, id,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Category,
		&p.SubType, &p.DurationDays, &p.SessionsCount, &p.IsActive, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

// ---- Subscription Repository ----

type subscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) repository.SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, clientID, productID int64) (*domain.ClientSubscription, error) {
	const q = `
		INSERT INTO client_subscriptions (client_id, product_id, start_date, end_date, sessions_left)
		SELECT $1, $2, NOW(),
		       NOW() + (p.duration_days || ' days')::interval,
		       p.sessions_count
		FROM products p WHERE p.id=$2
		RETURNING id, client_id, product_id, start_date, end_date, sessions_left, is_active, created_at`

	s := &domain.ClientSubscription{}
	err := r.db.QueryRow(ctx, q, clientID, productID).Scan(
		&s.ID, &s.ClientID, &s.ProductID,
		&s.StartDate, &s.EndDate, &s.SessionsLeft, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("subscriptionRepo.Create: %w", err)
	}
	return s, nil
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id int64) (*domain.ClientSubscription, error) {
	const q = `
		SELECT cs.id, cs.client_id, cs.product_id, cs.start_date, cs.end_date,
		       cs.sessions_left, cs.is_active, cs.created_at, p.name, p.price
		FROM client_subscriptions cs
		JOIN products p ON p.id = cs.product_id
		WHERE cs.id=$1`

	s := &domain.ClientSubscription{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.ClientID, &s.ProductID,
		&s.StartDate, &s.EndDate, &s.SessionsLeft, &s.IsActive, &s.CreatedAt,
		&s.ProductName, &s.Price,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return s, err
}

func (r *subscriptionRepo) ListByClient(ctx context.Context, clientID int64) ([]*domain.ClientSubscription, error) {
	const q = `
		SELECT cs.id, cs.client_id, cs.product_id, cs.start_date, cs.end_date,
		       cs.sessions_left, cs.is_active, cs.created_at, p.name, p.price
		FROM client_subscriptions cs
		JOIN products p ON p.id = cs.product_id
		WHERE cs.client_id=$1 ORDER BY cs.created_at DESC`

	rows, err := r.db.Query(ctx, q, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ClientSubscription
	for rows.Next() {
		s := &domain.ClientSubscription{}
		if err := rows.Scan(
			&s.ID, &s.ClientID, &s.ProductID,
			&s.StartDate, &s.EndDate, &s.SessionsLeft, &s.IsActive, &s.CreatedAt,
			&s.ProductName, &s.Price,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *subscriptionRepo) GetActiveByClient(ctx context.Context, clientID int64) (*domain.ClientSubscription, error) {
	const q = `
		SELECT cs.id, cs.client_id, cs.product_id, cs.start_date, cs.end_date,
		       cs.sessions_left, cs.is_active, cs.created_at, p.name, p.price
		FROM client_subscriptions cs
		JOIN products p ON p.id = cs.product_id
		WHERE cs.client_id=$1 AND cs.is_active=true AND cs.end_date > NOW()
		ORDER BY cs.end_date DESC LIMIT 1`

	s := &domain.ClientSubscription{}
	err := r.db.QueryRow(ctx, q, clientID).Scan(
		&s.ID, &s.ClientID, &s.ProductID,
		&s.StartDate, &s.EndDate, &s.SessionsLeft, &s.IsActive, &s.CreatedAt,
		&s.ProductName, &s.Price,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return s, err
}

func (r *subscriptionRepo) DecrementSessions(ctx context.Context, id int64) error {
	const q = `UPDATE client_subscriptions SET sessions_left = sessions_left - 1 WHERE id=$1 AND sessions_left > 0`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *subscriptionRepo) Deactivate(ctx context.Context, id int64) error {
	const q = `UPDATE client_subscriptions SET is_active=false WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
