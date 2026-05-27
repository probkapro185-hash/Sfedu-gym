package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — пользовательские данные в токене
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

// Manager — менеджер JWT токенов
type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

// Generate — создать токен
func (m *Manager) Generate(claims Claims, ttl time.Duration) (string, error) {
	jc := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: claims.UserID,
		Role:   claims.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jc)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt.Generate: %w", err)
	}
	return signed, nil
}

// Parse — разобрать и проверить токен
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt.Parse: %w", err)
	}

	jc, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt.Parse: invalid token")
	}
	return &Claims{UserID: jc.UserID, Role: jc.Role}, nil
}
