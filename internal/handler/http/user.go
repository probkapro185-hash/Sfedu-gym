package httphandler

import (
	"net/http"
	"strconv"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/middleware"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/service"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	user, err := h.userSvc.GetByID(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, user)
}

// PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	var input domain.UpdateUserInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.userSvc.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, user)
}

// PUT /api/v1/users/me/password
func (h *UserHandler) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	var input domain.ChangePasswordInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.userSvc.ChangePassword(r.Context(), userID, input); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/v1/users  — менеджер/админ: список клиентов
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repository.UserFilter{
		Search: q.Get("search"),
	}
	if roleStr := q.Get("role"); roleStr != "" {
		r := domain.Role(roleStr)
		filter.Role = &r
	}
	users, err := h.userSvc.List(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, users)
}

// POST /api/v1/users  — менеджер/админ: создать пользователя
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateUserInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.userSvc.CreateUser(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, user)
}

// GET /api/v1/users/{id}  — менеджер/админ
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, user)
}

// PUT /api/v1/users/{id}  — менеджер/админ: редактировать данные клиента
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var input domain.UpdateUserInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.userSvc.UpdateProfile(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, user)
}

// DELETE /api/v1/users/{id}  — только админ
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.userSvc.Delete(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusNoContent, nil)
}

// PATCH /api/v1/users/{id}/activate  — активировать
// PATCH /api/v1/users/{id}/deactivate — деактивировать
func (h *UserHandler) SetActive(active bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDFromPath(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		if err := h.userSvc.SetActive(r.Context(), id, active); err != nil {
			handleError(w, err)
			return
		}
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// GET /api/v1/applications  — менеджер/админ: заявки на регистрацию
func (h *UserHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	apps, err := h.userSvc.ListApplications(r.Context(), status)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, apps)
}

// POST /api/v1/applications/{id}/approve  — принять заявку
func (h *UserHandler) ApproveApplication(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil || body.Password == "" {
		respondError(w, http.StatusBadRequest, "password is required")
		return
	}
	user, err := h.userSvc.ApproveApplication(r.Context(), id, body.Password)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, user)
}

// POST /api/v1/applications/{id}/reject  — отклонить заявку
func (h *UserHandler) RejectApplication(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	if err := h.userSvc.RejectApplication(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// parseIDFromPath — получить числовой ID из последнего сегмента пути
func parseIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
