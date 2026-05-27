package httphandler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/middleware"
	"github.com/sfedu-crm/internal/service"
)

type ScheduleHandler struct {
	scheduleSvc *service.ScheduleService
}

func NewScheduleHandler(scheduleSvc *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleSvc: scheduleSvc}
}

// GET /api/v1/schedule?date_from=&date_to=&client_id=&trainer_id=
// Клиент видит только своё расписание; менеджер/админ — всё
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ScheduleFilter{}

	// Парсинг дат
	if from := q.Get("date_from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid date_from format (YYYY-MM-DD)")
			return
		}
		filter.DateFrom = t
	} else {
		// Если не указано — показать текущий месяц
		now := time.Now()
		filter.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	if to := q.Get("date_to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid date_to format (YYYY-MM-DD)")
			return
		}
		filter.DateTo = t
	} else {
		filter.DateTo = filter.DateFrom.AddDate(0, 1, 0)
	}

	// Клиент видит только своё расписание
	role, _ := middleware.GetRole(r.Context())
	if role == domain.RoleClient {
		userID, _ := middleware.GetUserID(r.Context())
		filter.ClientID = &userID
	} else if clientIDStr := q.Get("client_id"); clientIDStr != "" {
		// менеджер/админ может фильтровать по клиенту
		var cid int64
		if _, err := parseIntStr(clientIDStr, &cid); err == nil {
			filter.ClientID = &cid
		}
	}

	trainings, err := h.scheduleSvc.GetSchedule(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, trainings)
}

// GET /api/v1/schedule/{id}
func (h *ScheduleHandler) GetTraining(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid training id")
		return
	}
	training, err := h.scheduleSvc.GetTrainingByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, training)
}

// POST /api/v1/schedule/requests  — клиент подаёт заявку на тренировку
func (h *ScheduleHandler) SubmitTrainingRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	var input domain.CreateTrainingRequestInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req, err := h.scheduleSvc.SubmitTrainingRequest(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, req)
}

// GET /api/v1/schedule/requests  — менеджер/админ: все pending заявки
func (h *ScheduleHandler) ListPendingRequests(w http.ResponseWriter, r *http.Request) {
	list, err := h.scheduleSvc.ListPendingRequests(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

// GET /api/v1/schedule/requests/my  — клиент: свои заявки
func (h *ScheduleHandler) ListMyRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	list, err := h.scheduleSvc.ListClientRequests(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

// POST /api/v1/schedule/requests/{id}/approve  — менеджер/админ: принять заявку
func (h *ScheduleHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request id")
		return
	}
	var input domain.CreateTrainingInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	training, err := h.scheduleSvc.ApproveRequest(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, training)
}

func (h *ScheduleHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request id")
		return
	}
	if err := h.scheduleSvc.RejectRequest(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// PUT /api/v1/schedule/{id}  — менеджер/админ: перенести занятие
func (h *ScheduleHandler) UpdateTraining(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid training id")
		return
	}
	var input domain.UpdateTrainingInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	training, err := h.scheduleSvc.UpdateTraining(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, training)
}

// DELETE /api/v1/schedule/{id}  — только админ
func (h *ScheduleHandler) DeleteTraining(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid training id")
		return
	}
	if err := h.scheduleSvc.DeleteTraining(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusNoContent, nil)
}

func parseIntStr(s string, dst *int64) (int64, error) {
	if s == "" {
		return 0, domain.ErrInvalidInput
	}
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, domain.ErrInvalidInput
		}
		n = n*10 + int64(c-'0')
	}
	*dst = n
	return n, nil
}
