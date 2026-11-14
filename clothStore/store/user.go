package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	
)

var (
	ErrEmailExists  = errors.New("email is already registered")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	UserID       int    `json:"user_id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	PasswordHash string `json:"-"`
}

type UserStore struct {
	DB *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{DB: db}
}

func (s *UserStore) CreateUser(ctx context.Context, fullName, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO "User" (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id, full_name, email, role
	`
	var u User
	err := s.DB.QueryRow(ctx, query, fullName, email, passwordHash).Scan(
		&u.UserID,
		&u.FullName,
		&u.Email,
		&u.Role,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT user_id, full_name, email, role, password_hash
		FROM "User"
		WHERE email = $1
	`
	var u User
	err := s.DB.QueryRow(ctx, query, email).Scan(
		&u.UserID,
		&u.FullName,
		&u.Email,
		&u.Role,
		&u.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) UpdatePasswordByEmail(ctx context.Context, email, newHash string) error {
    query := `UPDATE users SET password_hash = $1 WHERE email = $2`
    tag, err := s.DB.Exec(ctx, query, newHash, email)
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return errors.New("user not found")
    }
    return nil
}