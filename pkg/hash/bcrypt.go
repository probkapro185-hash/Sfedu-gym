package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt — обёртка над bcrypt для хеширования паролей
type Bcrypt struct {
	cost int
}

func NewBcrypt(cost int) *Bcrypt {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &Bcrypt{cost: cost}
}

// Hash — хешировать пароль
func (b *Bcrypt) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.Hash: %w", err)
	}
	return string(hashed), nil
}

// Compare — сравнить пароль с хешем
func (b *Bcrypt) Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
