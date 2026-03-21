package jwtmanager

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Вот сюда потом можно будет воткнуть UserRole типа master или client или admin
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secretKey string
	ttl       time.Duration
}

func (m *Manager) GetTTL() time.Duration {
	return m.ttl
}
func NewManager(secretKey string, ttl time.Duration) *Manager {
	return &Manager{secretKey: secretKey, ttl: ttl}
}

func (m *Manager) NewToken(userID string, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), //Это jti
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (m *Manager) Validate(tokenString string) (*Claims, error) {
	// Парсим и валидируем токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
