package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ristep/smanzy_backend/internal/models"
)

// CustomClaims represents the custom claims in the JWT token
type CustomClaims struct {
	UserID uint     `json:"user_id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	secretKey string
}

// NewJWTService creates a new JWT service with the given secret key
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: secretKey,
	}
}

// TokenPair represents both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (js *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Extract role names from user roles
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	// Generate access token (short-lived: 15 minutes)
	accessToken, err := js.generateToken(user, roleNames, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (long-lived: 7 days)
	refreshToken, err := js.generateToken(user, roleNames, 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateToken is a helper function to create a JWT token with the given duration
func (js *JWTService) generateToken(user *models.User, roleNames []string, duration time.Duration) (string, error) {
	now := time.Now()
	expirationTime := now.Add(duration)

	claims := CustomClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Roles:  roleNames,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "um-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(js.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token, returning the claims or an error
func (js *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is the expected one
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(js.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateRefreshToken is an alias for ValidateToken used for refresh tokens
func (js *JWTService) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	return js.ValidateToken(tokenString)
}
