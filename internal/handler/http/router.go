package httphandler

import (
	"net/http"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/middleware"
	"github.com/sfedu-crm/pkg/jwt"
)

// Handlers — контейнер всех HTTP-хэндлеров
type Handlers struct {
	Auth     *AuthHandler
	User     *UserHandler
	Schedule *ScheduleHandler
	Finance  *FinanceHandler
	Shop     *ShopHandler
	Trainer  *TrainerHandler
	AI       *AIHandler
}

// NewRouter — регистрация всех маршрутов
// Используется net/http ServeMux (Go 1.22+), без сторонних фреймворков
func NewRouter(h *Handlers, tokenMgr *jwt.Manager) http.Handler {
	mux := http.NewServeMux()

	// ==================== Публичные маршруты ====================
	// Форма «Записаться» (лендинг)
	mux.HandleFunc("POST /api/v1/auth/apply", h.Auth.SubmitApplication)
	// Вход в систему
	mux.HandleFunc("POST /api/v1/auth/login", h.Auth.Login)

	// Публичный каталог тренеров (лендинг)
	mux.HandleFunc("GET /api/v1/trainers", h.Trainer.ListTrainers)
	mux.HandleFunc("GET /api/v1/trainers/{id}", h.Trainer.GetTrainer)

	// ==================== Маршруты с JWT ====================
	authMw := middleware.Auth(tokenMgr)

	adminOnly := chain(authMw, middleware.RequireRole(domain.RoleAdmin))
	managerAndAdmin := chain(authMw, middleware.RequireRole(domain.RoleManager, domain.RoleAdmin))
	allRoles := chain(authMw, middleware.RequireRole(domain.RoleClient, domain.RoleManager, domain.RoleAdmin))

	// --- Профиль (все роли) ---
	mux.Handle("GET /api/v1/users/me", allRoles(http.HandlerFunc(h.User.GetMe)))
	mux.Handle("PUT /api/v1/users/me", allRoles(http.HandlerFunc(h.User.UpdateMe)))
	mux.Handle("PUT /api/v1/users/me/password", allRoles(http.HandlerFunc(h.User.ChangeMyPassword)))

	// --- Управление пользователями (менеджер + админ) ---
	mux.Handle("GET /api/v1/users", managerAndAdmin(http.HandlerFunc(h.User.ListUsers)))
	mux.Handle("POST /api/v1/users", managerAndAdmin(http.HandlerFunc(h.User.CreateUser)))
	mux.Handle("GET /api/v1/users/{id}", managerAndAdmin(http.HandlerFunc(h.User.GetUser)))
	mux.Handle("PUT /api/v1/users/{id}", managerAndAdmin(http.HandlerFunc(h.User.UpdateUser)))

	// --- Управление пользователями (только админ) ---
	mux.Handle("DELETE /api/v1/users/{id}", adminOnly(http.HandlerFunc(h.User.DeleteUser)))
	mux.Handle("PATCH /api/v1/users/{id}/activate", adminOnly(h.User.SetActive(true)))
	mux.Handle("PATCH /api/v1/users/{id}/deactivate", adminOnly(h.User.SetActive(false)))

	// --- Заявки на регистрацию (менеджер + админ) ---
	mux.Handle("GET /api/v1/applications", managerAndAdmin(http.HandlerFunc(h.User.ListApplications)))
	mux.Handle("POST /api/v1/applications/{id}/approve", managerAndAdmin(http.HandlerFunc(h.User.ApproveApplication)))
	mux.Handle("POST /api/v1/applications/{id}/reject", managerAndAdmin(http.HandlerFunc(h.User.RejectApplication)))

	// --- Расписание ---
	// Клиент: своё расписание и свои заявки
	mux.Handle("GET /api/v1/schedule", allRoles(http.HandlerFunc(h.Schedule.GetSchedule)))
	mux.Handle("GET /api/v1/schedule/{id}", allRoles(http.HandlerFunc(h.Schedule.GetTraining)))
	mux.Handle("POST /api/v1/schedule/requests", allRoles(http.HandlerFunc(h.Schedule.SubmitTrainingRequest)))
	mux.Handle("GET /api/v1/schedule/requests/my", allRoles(http.HandlerFunc(h.Schedule.ListMyRequests)))

	// Менеджер/Админ: управление заявками и занятиями
	mux.Handle("GET /api/v1/schedule/requests", managerAndAdmin(http.HandlerFunc(h.Schedule.ListPendingRequests)))
	mux.Handle("POST /api/v1/schedule/requests/{id}/approve", managerAndAdmin(http.HandlerFunc(h.Schedule.ApproveRequest)))
	mux.Handle("POST /api/v1/schedule/requests/{id}/reject", managerAndAdmin(http.HandlerFunc(h.Schedule.RejectRequest)))
	mux.Handle("PUT /api/v1/schedule/{id}", managerAndAdmin(http.HandlerFunc(h.Schedule.UpdateTraining)))
	mux.Handle("DELETE /api/v1/schedule/{id}", adminOnly(http.HandlerFunc(h.Schedule.DeleteTraining)))

	// --- Финансы ---
	// Клиент: свой баланс и история
	mux.Handle("POST /api/v1/finance/me/topup", allRoles(http.HandlerFunc(h.Finance.TopUpMyBalance)))
	mux.Handle("GET /api/v1/finance/me/payments", allRoles(http.HandlerFunc(h.Finance.GetMyPayments)))

	// Менеджер: пополнение баланса клиента
	mux.Handle("POST /api/v1/finance/topup", managerAndAdmin(http.HandlerFunc(h.Finance.TopUpBalance)))

	// Только Админ: полная финансовая история и сводка
	mux.Handle("GET /api/v1/finance/payments", adminOnly(http.HandlerFunc(h.Finance.ListPayments)))
	mux.Handle("GET /api/v1/finance/summary", adminOnly(http.HandlerFunc(h.Finance.GetSummary)))

	// --- Магазин ---
	mux.Handle("GET /api/v1/shop/products", allRoles(http.HandlerFunc(h.Shop.ListProducts)))
	mux.Handle("GET /api/v1/shop/products/{id}", allRoles(http.HandlerFunc(h.Shop.GetProduct)))
	mux.Handle("POST /api/v1/shop/purchase", allRoles(http.HandlerFunc(h.Shop.PurchaseProduct)))
	mux.Handle("GET /api/v1/shop/my-subscriptions", allRoles(http.HandlerFunc(h.Shop.GetMySubscriptions)))

	// Менеджер/Админ: управление товарами
	mux.Handle("POST /api/v1/shop/products", managerAndAdmin(http.HandlerFunc(h.Shop.CreateProduct)))
	mux.Handle("PUT /api/v1/shop/products/{id}", managerAndAdmin(http.HandlerFunc(h.Shop.UpdateProduct)))
	mux.Handle("DELETE /api/v1/shop/products/{id}", adminOnly(http.HandlerFunc(h.Shop.DeleteProduct)))

	// --- Тренеры (управление — только админ) ---
	mux.Handle("POST /api/v1/trainers", adminOnly(http.HandlerFunc(h.Trainer.CreateTrainer)))
	mux.Handle("PUT /api/v1/trainers/{id}", adminOnly(http.HandlerFunc(h.Trainer.UpdateTrainer)))
	mux.Handle("DELETE /api/v1/trainers/{id}", adminOnly(http.HandlerFunc(h.Trainer.DeleteTrainer)))

	// --- ИИ Спортсмен (все авторизованные) ---
	mux.Handle("POST /api/v1/ai/chat", allRoles(http.HandlerFunc(h.AI.Chat)))

	return mux
}

// chain — объединить middleware в цепочку
func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
