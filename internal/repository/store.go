package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *sqlx.DB {
	return s.db
}

func (s *Store) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `insert into users (id, email, phone, full_name, role, password_hash, photo_file_id, is_active, created_at, updated_at)
values (:id, :email, :phone, :full_name, :role, :password_hash, :photo_file_id, :is_active, :created_at, :updated_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return domain.User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return domain.User{}, err
		}
	}
	return user, rows.Err()
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var user domain.User
	err := s.db.GetContext(ctx, &user, `select * from users where id = $1`, id)
	return user, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := s.db.GetContext(ctx, &user, `select * from users where email = $1`, email)
	return user, err
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := s.db.SelectContext(ctx, &users, `select * from users order by created_at desc`)
	return users, err
}

func (s *Store) UpdateUser(ctx context.Context, id uuid.UUID, phone *string, fullName *string, isActive *bool) (domain.User, error) {
	var user domain.User
	err := s.db.GetContext(ctx, &user, `
update users set phone = coalesce($2, phone), full_name = coalesce($3, full_name),
is_active = coalesce($4, is_active), updated_at = now()
where id = $1 returning *`, id, phone, fullName, isActive)
	return user, err
}

func (s *Store) CreatePlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error) {
	const query = `insert into subscription_plans (id, name, description, duration_days, price, currency, is_active, created_at, updated_at)
values (:id, :name, :description, :duration_days, :price, :currency, :is_active, :created_at, :updated_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, plan)
	if err != nil {
		return domain.SubscriptionPlan{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&plan); err != nil {
			return domain.SubscriptionPlan{}, err
		}
	}
	return plan, rows.Err()
}

func (s *Store) GetPlan(ctx context.Context, id uuid.UUID) (domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := s.db.GetContext(ctx, &plan, `select * from subscription_plans where id = $1`, id)
	return plan, err
}

func (s *Store) ListPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	var plans []domain.SubscriptionPlan
	err := s.db.SelectContext(ctx, &plans, `select * from subscription_plans order by created_at desc`)
	return plans, err
}

func (s *Store) UpdatePlan(ctx context.Context, id uuid.UUID, name *string, description *string, durationDays *int, price *decimal.Decimal, currency *string, isActive *bool) (domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := s.db.GetContext(ctx, &plan, `
update subscription_plans set name = coalesce($2, name), description = coalesce($3, description),
duration_days = coalesce($4, duration_days), price = coalesce($5, price), currency = coalesce($6, currency),
is_active = coalesce($7, is_active), updated_at = now()
where id = $1 returning *`, id, name, description, durationDays, price, currency, isActive)
	return plan, err
}

func (s *Store) CreateSubscription(ctx context.Context, sub domain.UserSubscription) (domain.UserSubscription, error) {
	const query = `insert into user_subscriptions (id, user_id, plan_id, starts_at, ends_at, status, created_at, updated_at)
values (:id, :user_id, :plan_id, :starts_at, :ends_at, :status, :created_at, :updated_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, sub)
	if err != nil {
		return domain.UserSubscription{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&sub); err != nil {
			return domain.UserSubscription{}, err
		}
	}
	return sub, rows.Err()
}

func (s *Store) GetSubscription(ctx context.Context, id uuid.UUID) (domain.UserSubscription, error) {
	var sub domain.UserSubscription
	err := s.db.GetContext(ctx, &sub, `select * from user_subscriptions where id = $1`, id)
	return sub, err
}

func (s *Store) ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserSubscription, error) {
	var subs []domain.UserSubscription
	err := s.db.SelectContext(ctx, &subs, `select * from user_subscriptions where user_id = $1 order by created_at desc`, userID)
	return subs, err
}

func (s *Store) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status domain.SubscriptionStatus) (domain.UserSubscription, error) {
	var sub domain.UserSubscription
	err := s.db.GetContext(ctx, &sub, `update user_subscriptions set status = $2, updated_at = now() where id = $1 returning *`, id, status)
	return sub, err
}

func (s *Store) CreateQRPass(ctx context.Context, pass domain.QRPass) (domain.QRPass, error) {
	const query = `insert into qr_passes (id, user_id, subscription_id, token_hash, status, expires_at, created_at)
values (:id, :user_id, :subscription_id, :token_hash, :status, :expires_at, :created_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, pass)
	if err != nil {
		return domain.QRPass{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&pass); err != nil {
			return domain.QRPass{}, err
		}
	}
	return pass, rows.Err()
}

func (s *Store) RevokeQRPass(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `update qr_passes set status = 'revoked' where id = $1`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetLatestQRPassByUser(ctx context.Context, userID uuid.UUID) (domain.QRPass, error) {
	var pass domain.QRPass
	err := s.db.GetContext(ctx, &pass, `select * from qr_passes where user_id = $1 order by created_at desc limit 1`, userID)
	return pass, err
}

type QRValidationRecord struct {
	Pass         domain.QRPass
	User         domain.User
	Subscription domain.UserSubscription
}

func (s *Store) GetQRValidationRecord(ctx context.Context, tokenHash string) (QRValidationRecord, error) {
	var row struct {
		PassID        uuid.UUID                 `db:"pass_id"`
		PassUserID    uuid.UUID                 `db:"pass_user_id"`
		PassSubID     uuid.UUID                 `db:"pass_subscription_id"`
		TokenHash     string                    `db:"token_hash"`
		PassStatus    domain.QRPassStatus       `db:"pass_status"`
		PassExpiresAt time.Time                 `db:"pass_expires_at"`
		PassCreatedAt time.Time                 `db:"pass_created_at"`
		UserID        uuid.UUID                 `db:"user_id"`
		Email         string                    `db:"email"`
		Phone         *string                   `db:"phone"`
		FullName      string                    `db:"full_name"`
		Role          domain.Role               `db:"role"`
		PasswordHash  *string                   `db:"password_hash"`
		PhotoFileID   *uuid.UUID                `db:"photo_file_id"`
		IsActive      bool                      `db:"is_active"`
		UserCreatedAt time.Time                 `db:"user_created_at"`
		UserUpdatedAt time.Time                 `db:"user_updated_at"`
		SubID         uuid.UUID                 `db:"sub_id"`
		SubUserID     uuid.UUID                 `db:"sub_user_id"`
		PlanID        uuid.UUID                 `db:"plan_id"`
		StartsAt      time.Time                 `db:"starts_at"`
		EndsAt        time.Time                 `db:"ends_at"`
		SubStatus     domain.SubscriptionStatus `db:"sub_status"`
		SubCreatedAt  time.Time                 `db:"sub_created_at"`
		SubUpdatedAt  time.Time                 `db:"sub_updated_at"`
	}
	err := s.db.GetContext(ctx, &row, `
select q.id as pass_id, q.user_id as pass_user_id, q.subscription_id as pass_subscription_id,
q.token_hash, q.status as pass_status, q.expires_at as pass_expires_at, q.created_at as pass_created_at,
u.id as user_id, u.email, u.phone, u.full_name, u.role, u.password_hash, u.photo_file_id,
u.is_active, u.created_at as user_created_at, u.updated_at as user_updated_at,
us.id as sub_id, us.user_id as sub_user_id, us.plan_id, us.starts_at, us.ends_at,
us.status as sub_status, us.created_at as sub_created_at, us.updated_at as sub_updated_at
from qr_passes q
join users u on u.id = q.user_id
join user_subscriptions us on us.id = q.subscription_id
where q.token_hash = $1`, tokenHash)
	if err != nil {
		return QRValidationRecord{}, err
	}
	return QRValidationRecord{
		Pass:         domain.QRPass{ID: row.PassID, UserID: row.PassUserID, SubscriptionID: row.PassSubID, TokenHash: row.TokenHash, Status: row.PassStatus, ExpiresAt: row.PassExpiresAt, CreatedAt: row.PassCreatedAt},
		User:         domain.User{ID: row.UserID, Email: row.Email, Phone: row.Phone, FullName: row.FullName, Role: row.Role, PasswordHash: row.PasswordHash, PhotoFileID: row.PhotoFileID, IsActive: row.IsActive, CreatedAt: row.UserCreatedAt, UpdatedAt: row.UserUpdatedAt},
		Subscription: domain.UserSubscription{ID: row.SubID, UserID: row.SubUserID, PlanID: row.PlanID, StartsAt: row.StartsAt, EndsAt: row.EndsAt, Status: row.SubStatus, CreatedAt: row.SubCreatedAt, UpdatedAt: row.SubUpdatedAt},
	}, nil
}

func (s *Store) CreateAccessEvent(ctx context.Context, event domain.AccessEvent) (domain.AccessEvent, error) {
	const query = `insert into access_events (id, user_id, subscription_id, qr_pass_id, event_type, decision, reason, scanner_id, photo_file_id, occurred_at, created_at)
values (:id, :user_id, :subscription_id, :qr_pass_id, :event_type, :decision, :reason, :scanner_id, :photo_file_id, :occurred_at, :created_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, event)
	if err != nil {
		return domain.AccessEvent{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&event); err != nil {
			return domain.AccessEvent{}, err
		}
	}
	return event, rows.Err()
}

func (s *Store) LatestAllowedAccessEvent(ctx context.Context, userID uuid.UUID) (*domain.AccessEvent, error) {
	var event domain.AccessEvent
	err := s.db.GetContext(ctx, &event, `select * from access_events where user_id = $1 and decision = 'allowed' order by occurred_at desc limit 1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *Store) ListAccessEvents(ctx context.Context) ([]domain.AccessEvent, error) {
	var events []domain.AccessEvent
	err := s.db.SelectContext(ctx, &events, `select * from access_events order by occurred_at desc limit 1000`)
	return events, err
}

func (s *Store) CreatePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	const query = `insert into payments (id, user_id, subscription_id, amount, currency, method, status, receipt_file_id, approved_by, approved_at, created_at, updated_at)
values (:id, :user_id, :subscription_id, :amount, :currency, :method, :status, :receipt_file_id, :approved_by, :approved_at, :created_at, :updated_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, payment)
	if err != nil {
		return domain.Payment{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&payment); err != nil {
			return domain.Payment{}, err
		}
	}
	return payment, rows.Err()
}

func (s *Store) ListPayments(ctx context.Context) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := s.db.SelectContext(ctx, &payments, `select * from payments order by created_at desc`)
	return payments, err
}

func (s *Store) GetPayment(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	var payment domain.Payment
	err := s.db.GetContext(ctx, &payment, `select * from payments where id = $1`, id)
	return payment, err
}

func (s *Store) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, approvedBy *uuid.UUID) (domain.Payment, error) {
	var payment domain.Payment
	var approvedAt *time.Time
	if status == domain.PaymentApproved {
		now := time.Now().UTC()
		approvedAt = &now
	}
	err := s.db.GetContext(ctx, &payment, `update payments set status = $2, approved_by = $3, approved_at = $4, updated_at = now() where id = $1 returning *`, id, status, approvedBy, approvedAt)
	return payment, err
}

func (s *Store) CreateFile(ctx context.Context, file domain.File) (domain.File, error) {
	const query = `insert into files (id, bucket, object_key, original_name, content_type, size_bytes, created_at)
values (:id, :bucket, :object_key, :original_name, :content_type, :size_bytes, :created_at) returning *`
	rows, err := s.db.NamedQueryContext(ctx, query, file)
	if err != nil {
		return domain.File{}, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&file); err != nil {
			return domain.File{}, err
		}
	}
	return file, rows.Err()
}
