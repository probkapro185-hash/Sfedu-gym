package domain

import "errors"

// Sentinel errors used across the application
var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidPhone     = errors.New("invalid phone number")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// AIAssistantMessage — структура для будущего ИИ-ассистента «ИИ Спортсмен»
// Реализация ИИ будет добавлена позже; здесь только контракт
type AIAssistantMessage struct {
	Role    string `json:"role"`    // "user" | "assistant"
	Content string `json:"content"`
}

type AIAssistantRequest struct {
	Messages []AIAssistantMessage `json:"messages"`
	UserID   int64                `json:"user_id"`
}

type AIAssistantResponse struct {
	Reply string `json:"reply"`
}
