package service

import (
	"context"

	"github.com/sfedu-crm/internal/domain"
)

// AIAssistantService — заглушка для будущего ИИ-ассистента «ИИ Спортсмен»
// Когда придёт время — реализовать интеграцию с LLM (OpenAI/Anthropic/etc.)
type AIAssistantService struct{}

func NewAIAssistantService() *AIAssistantService {
	return &AIAssistantService{}
}

// Chat — обработка диалога с ИИ-ассистентом
// Сейчас возвращает заглушку; в будущем — запрос к LLM API
func (s *AIAssistantService) Chat(_ context.Context, req domain.AIAssistantRequest) (*domain.AIAssistantResponse, error) {
	return &domain.AIAssistantResponse{
		Reply: "ИИ Спортсмен временно недоступен. Скоро буду помогать вам с навигацией на сайте!",
	}, nil
}
