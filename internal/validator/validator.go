package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/sfedu-crm/internal/domain"
)

// Российские номера: +7(987)654-32-10 / +79871234567 / 89871234567
var phoneRegex = regexp.MustCompile(`^(\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}$`)

// Email: только @gmail.com и @mail.ru (как указано в ТЗ)
var emailRegex = regexp.MustCompile(`(?i)^[a-zA-Z0-9._%+\-]+@(gmail\.com|mail\.ru)$`)

// ValidatePhone — валидация российского номера телефона
func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("%w: must be a valid Russian phone number", domain.ErrInvalidPhone)
	}
	return nil
}

// ValidateEmail — валидация email (только gmail.com и mail.ru)
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: only @gmail.com and @mail.ru are allowed", domain.ErrInvalidEmail)
	}
	return nil
}

// ValidateFullName — ФИО: минимум 2 слова, только кириллица и пробелы
func ValidateFullName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: full name is required", domain.ErrInvalidInput)
	}
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return fmt.Errorf("%w: full name must contain at least first and last name", domain.ErrInvalidInput)
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' {
			return fmt.Errorf("%w: full name must contain only letters", domain.ErrInvalidInput)
		}
	}
	return nil
}

// ValidatePassword — минимальные требования к паролю
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}
	return nil
}
