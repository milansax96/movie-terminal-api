package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/milansax96/movie-terminal-api/config"
	"github.com/milansax96/movie-terminal-api/internal/middleware"
	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/repository"

	"gorm.io/gorm"
)

// AuthService handles authentication via Google OAuth.
type AuthService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

// AuthResult holds the result of a login attempt.
type AuthResult struct {
	User  *models.User
	Token string
	IsNew bool
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *AuthService) fetchGoogleUserInfo(ctx context.Context, accessToken string) (_ *googleUserInfo, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling userinfo endpoint: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidToken
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding userinfo: %w", err)
	}

	return &info, nil
}

// GoogleLogin authenticates a user via Google OAuth access token.
func (s *AuthService) GoogleLogin(ctx context.Context, accessToken string) (*AuthResult, error) {
	info, err := s.fetchGoogleUserInfo(ctx, accessToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	googleID := info.Sub
	email := info.Email
	name := info.Name
	picture := info.Picture

	if googleID == "" || email == "" {
		return nil, ErrMissingClaims
	}

	user, err := s.userRepo.FindByGoogleID(googleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		username := name
		if username == "" {
			username = email
		}

		user = &models.User{
			Username:       username,
			Email:          email,
			GoogleID:       googleID,
			ProfilePicture: picture,
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, ErrAlreadyExists
		}

		token, err := s.generateToken(user.ID.String())
		if err != nil {
			return nil, err
		}

		return &AuthResult{User: user, Token: token, IsNew: true}, nil
	} else if err != nil {
		return nil, err
	}

	// Update profile picture if changed
	if picture != "" && picture != user.ProfilePicture {
		err := s.userRepo.UpdateProfilePicture(user.ID, picture)
		if err != nil {
			return nil, err
		}

		user.ProfilePicture = picture
	}

	token, err := s.generateToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Token: token, IsNew: false}, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}
