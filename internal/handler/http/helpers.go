package httphandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sfedu-crm/internal/domain"
)

// respond — сериализовать JSON-ответ
func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// respondError — стандартный формат ошибки
func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}

// decode — десериализовать тело запроса
func decode(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// handleError — маппинг domain-ошибок в HTTP статусы
func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrInvalidPhone),
		errors.Is(err, domain.ErrInvalidEmail):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrInvalidPassword):
		respondError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		respondError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInsufficientFunds):
		respondError(w, http.StatusPaymentRequired, err.Error())
	default:
		// Выводим реальную ошибку в лог и в ответ (в dev-режиме удобно)
		slog.Error("unhandled error", "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}
