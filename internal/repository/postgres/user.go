package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jordanhimawan/payroll-mgmt/internal/models"
	"github.com/jordanhimawan/payroll-mgmt/internal/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, role, salary, is_active, created_at, updated_at, created_by, updated_by
		FROM users 
		WHERE username = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.Salary, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&user.CreatedBy, &user.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, role, salary, is_active, created_at, updated_at, created_by, updated_by
		FROM users 
		WHERE id = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.Salary, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&user.CreatedBy, &user.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, role, salary, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.Role, user.Salary, user.CreatedBy,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}
