package userrepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/user"
)

//go:generate mockgen -source=user_repo.go -destination=user_repo_mock.go -package=userrepository
type UserRepository interface {
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (*user.User, error)
	FindUserByID(ctx context.Context, userID string) (*user.User, error)
	ValidateRefreshToken(ctx context.Context, token string) error

	// Transaction
	InsertUserTx(ctx context.Context, tx *sql.Tx, u *user.User) error
	InsertRefreshTokenTx(ctx context.Context, tx *sql.Tx, token *user.RefreshToken) error
	RevokedRefreshTokenTx(ctx context.Context, tx *sql.Tx, token string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) InsertUserTx(ctx context.Context, tx *sql.Tx, u *user.User) error {
	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2) RETURNING id, created_at, updated_at
	`
	if err := tx.QueryRowContext(ctx, query, u.Email, u.Password).Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (r *userRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var dummy bool
	query := `SELECT 1 FROM users WHERE email = $1 LIMIT 1`

	if err := r.db.QueryRowContext(ctx, query, email).Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *userRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	query := `
		SELECT id, email, password
		FROM users WHERE email = $1 LIMIT 1
	`
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindUserByID(ctx context.Context, userID string) (*user.User, error) {
	var u user.User
	query := `
		SELECT id, email, created_at, updated_at
		FROM users WHERE id = $1 LIMIT 1
	`
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) InsertRefreshTokenTx(ctx context.Context, tx *sql.Tx, token *user.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.Revoked,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ValidateRefreshToken(ctx context.Context, token string) error {
	var revoked bool
	var expiresAt time.Time

	query := `
		SELECT revoked, expires_at
		FROM refresh_tokens WHERE token = $1 LIMIT 1
	`
	if err := r.db.QueryRowContext(ctx, query, token).Scan(
		&revoked,
		&expiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrTokenNotFound
		}
		return err
	}

	if revoked {
		return errs.ErrTokenRevoked
	}

	if time.Now().After(expiresAt) {
		return errs.ErrTokenExpires
	}
	return nil
}

func (r *userRepository) RevokedRefreshTokenTx(ctx context.Context, tx *sql.Tx, token string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`
	res, err := tx.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errs.ErrTokenNotFound
	}
	return nil
}
