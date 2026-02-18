package repository

import (
	"context"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (email, password, name, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	err := r.db.QueryRow(context.Background(), query, user.Email, user.Password, user.Name).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (r *postgresUserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password, name, created_at FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *postgresUserRepository) GetByID(id int64) (*domain.User, error) {
	query := `SELECT id, email, password, name, created_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}
