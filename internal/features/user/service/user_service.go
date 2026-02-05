package userservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/user"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/codepnw/go-starter-kit/pkg/utils/password"
)

type UserService interface {
	Register(ctx context.Context, u *user.User) (*UserTokenResponse, error)
	Login(ctx context.Context, email, password string) (*UserTokenResponse, error)
	RefreshToken(ctx context.Context, token string) (*UserTokenResponse, error)
	Logout(ctx context.Context, token string) error
	GetProfile(ctx context.Context) (*user.User, error)
}

type userService struct {
	tx    database.TxManager
	token jwttoken.JWTToken
	repo  userrepository.UserRepository
}

func NewUserService(tx database.TxManager, token jwttoken.JWTToken, repo userrepository.UserRepository) UserService {
	return &userService{
		tx:    tx,
		token: token,
		repo:  repo,
	}
}

type UserTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *userService) Register(ctx context.Context, u *user.User) (*UserTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Check Email Exists
	exists, err := s.repo.CheckEmailExists(ctx, u.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrEmailAlreadyExists
	}

	// Hash Password
	hashedPassword, err := password.GenerateHashPassword(u.Password)
	if err != nil {
		return nil, err
	}
	u.Password = hashedPassword

	var response *UserTokenResponse
	// DB Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Insert User
		if err := s.repo.InsertUserTx(ctx, tx, u); err != nil {
			return err
		}

		// Generate Token
		resp, err := s.generateToken(u)
		if err != nil {
			return err
		}

		// Save Refresh Token
		insertTokenInput := s.insertRefreshTokenInput(u.ID, resp.RefreshToken)
		if err := s.repo.InsertRefreshTokenTx(ctx, tx, insertTokenInput); err != nil {
			return err
		}

		response = resp
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *userService) Login(ctx context.Context, email string, pwd string) (*UserTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Find User Email
	foundUser, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, errs.ErrInvalidEmailOrPassword
	}

	// Verify Password
	if ok := password.CompareHashedPassword(foundUser.Password, pwd); !ok {
		return nil, errs.ErrInvalidEmailOrPassword
	}

	var response *UserTokenResponse
	// DB Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Generate Token
		resp, err := s.generateToken(foundUser)
		if err != nil {
			return err
		}

		// Save Refresh Token
		insertTokenInput := s.insertRefreshTokenInput(foundUser.ID, resp.RefreshToken)
		if err := s.repo.InsertRefreshTokenTx(ctx, tx, insertTokenInput); err != nil {
			return err
		}

		response = resp
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *userService) RefreshToken(ctx context.Context, token string) (*UserTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Validate Token
	if err := s.repo.ValidateRefreshToken(ctx, token); err != nil {
		return nil, err
	}

	// Get UserID From Context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userData, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var response *UserTokenResponse
	// DB Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Revoked Old Token
		if err := s.repo.RevokedRefreshTokenTx(ctx, tx, token); err != nil {
			return err
		}

		// Generate New Token
		resp, err := s.generateToken(userData)
		if err != nil {
			return err
		}

		// Save New Token
		insertTokenInput := s.insertRefreshTokenInput(userData.ID, resp.RefreshToken)
		if err := s.repo.InsertRefreshTokenTx(ctx, tx, insertTokenInput); err != nil {
			return err
		}

		response = resp
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *userService) Logout(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	err := s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.RevokedRefreshTokenTx(ctx, tx, token); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) GetProfile(ctx context.Context) (*user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userData, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userData, nil
}

// ------------------ Private Method -------------------

func (s *userService) generateToken(u *user.User) (*UserTokenResponse, error) {
	accessToken, err := s.token.GenerateAccessToken(u)
	if err != nil {
		return nil, fmt.Errorf("failed gen access token: %w", err)
	}

	refreshToken, err := s.token.GenerateRefreshToken(u)
	if err != nil {
		return nil, fmt.Errorf("failed gen refresh token: %w", err)
	}

	response := &UserTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return response, nil
}

func (s *userService) insertRefreshTokenInput(userID, token string) *user.RefreshToken {
	return &user.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(config.RefreshTokenDuration),
		Revoked:   false,
	}
}
