package internal

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey []byte
}

type Role string

const (
	AdminRole  Role = "admin"
	ClientRole Role = "client"
)

type Claims struct {
	UserID int  `json:"userID"`
	Role   Role `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

type TokenAdminResponse struct {
	AccessToken  string `json:"accessAdminToken"`
	RefreshToken string `json:"refreshAdminToken"`
}

type TokenClienteResponse struct {
	AccessToken  string `json:"accessClientToken"`
	RefreshToken string `json:"refreshClientToken"`
}

type TokenType int

const (
	AccessToken TokenType = iota
	RefreshToken
)

func (s *JWTService) generateTokenPair(userID int, role Role) (TokenAdminResponse, TokenClienteResponse, error) {
	accessToken, err := s.generateToken(userID, role, AccessToken)
	if err != nil {
		return TokenAdminResponse{}, TokenClienteResponse{}, err
	}

	refreshToken, err := s.generateToken(userID, role, RefreshToken)
	if err != nil {
		return TokenAdminResponse{}, TokenClienteResponse{}, err
	}

	if role == AdminRole {
		return TokenAdminResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, TokenClienteResponse{}, nil
	} else if role == ClientRole {
		return TokenAdminResponse{}, TokenClienteResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}

	return TokenAdminResponse{}, TokenClienteResponse{}, errors.New("invalid role")
}

func (s *JWTService) GenerateTokenPairAdmin(userID int) (TokenAdminResponse, error) {
	tokens, _, err := s.generateTokenPair(userID, AdminRole)
	return tokens, err
}

func (s *JWTService) GenerateTokenPairClient(userID int) (TokenClienteResponse, error) {
	_, tokens, err := s.generateTokenPair(userID, ClientRole)
	return tokens, err
}

func (s *JWTService) generateToken(userID int, role Role, tokenType TokenType) (string, error) {
	var expTime time.Duration
	switch tokenType {
	case AccessToken:
		expTime = 24 * time.Hour
	case RefreshToken:
		expTime = 7 * 24 * time.Hour
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (int, Role, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, claims.Role, nil
	}

	return 0, "", errors.New("invalid token")
}
