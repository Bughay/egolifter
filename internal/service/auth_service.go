// internal/service/auth_service.go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bughay/go-backend-layout/internal/auth"
	"github.com/Bughay/go-backend-layout/internal/model"
	"github.com/Bughay/go-backend-layout/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the contract for authentication business logic.
type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.RegistrationResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, string, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*model.RefreshResponse, string, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.Manager
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.Manager) AuthService {
	return &authService{userRepo: userRepo, jwtManager: jwtManager}
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.RegistrationResponse, error) {
	// --- Validation ---
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, fmt.Errorf("validation: a valid email address is required")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("validation: password must be at least 8 characters long")
	}

	// --- Check for existing user ---
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("register: failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("validation: a user with this email already exists")
	}

	// --- Hash the password using bcrypt at cost 12 ---
	// Cost 12 is the modern recommendation: secure enough, not so slow it impacts UX.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("register: failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, req.Email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("register: failed to create user: %w", err)
	}

	if err != nil {
		return nil, err
	}

	return &model.RegistrationResponse{
		Message: "You have succesfully registered",
		Success: true,
		Person:  user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, string, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("login: database error: %w", err)
	}

	// IMPORTANT: Return the same generic error for both "user not found" and
	// "wrong password" to prevent user enumeration attacks.
	if user == nil {
		return nil, "", fmt.Errorf("auth: invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", fmt.Errorf("auth: invalid email or password")
	}

	accessToken, err := s.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", err
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		User:        user,
	}, refreshToken, nil
}

// Refresh validates the refresh token and issues NEW tokens (rotation)
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*model.RefreshResponse, string, error) {
	claims, err := s.jwtManager.Validate(refreshToken)
	if err != nil {
		return nil, "", fmt.Errorf("auth: invalid refresh token: %w", err)
	}

	// Verify user still exists
	user, err := s.userRepo.FindByEmail(ctx, claims.Email)
	if err != nil {
		return nil, "", fmt.Errorf("refresh: database error: %w", err)
	}
	if user == nil {
		return nil, "", fmt.Errorf("auth: user no longer exists")
	}

	// Issue NEW tokens (rotation — old refresh token is now invalid)
	newAccessToken, err := s.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", err
	}
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", err
	}

	return &model.RefreshResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
	}, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// Validate the token is even real before accepting logout
	_, err := s.jwtManager.Validate(refreshToken)
	if err != nil {
		return fmt.Errorf("auth: invalid token")
	}
	return nil
}
