package httphandler

import (
	"net/http"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/service"
)

// ---- Trainer Handler ----

type TrainerHandler struct {
	trainerSvc *service.TrainerService
}

func NewTrainerHandler(trainerSvc *service.TrainerService) *TrainerHandler {
	return &TrainerHandler{trainerSvc: trainerSvc}
}

// GET /api/v1/trainers?specialization=&search=
func (h *TrainerHandler) ListTrainers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repository.TrainerFilter{
		Search: q.Get("search"),
	}
	if spec := q.Get("specialization"); spec != "" {
		s := domain.TrainerSpecialization(spec)
		filter.Specialization = &s
	}
	active := true
	filter.IsActive = &active

	trainers, err := h.trainerSvc.List(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, trainers)
}

// GET /api/v1/trainers/{id}
func (h *TrainerHandler) GetTrainer(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trainer id")
		return
	}
	trainer, err := h.trainerSvc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, trainer)
}

// POST /api/v1/trainers  — только админ
func (h *TrainerHandler) CreateTrainer(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateTrainerInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	trainer, err := h.trainerSvc.Create(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, trainer)
}

// PUT /api/v1/trainers/{id}  — только админ
func (h *TrainerHandler) UpdateTrainer(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trainer id")
		return
	}
	var input domain.UpdateTrainerInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	trainer, err := h.trainerSvc.Update(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, trainer)
}

// DELETE /api/v1/trainers/{id}  — только админ
func (h *TrainerHandler) DeleteTrainer(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trainer id")
		return
	}
	if err := h.trainerSvc.Delete(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusNoContent, nil)
}

// ---- AI Assistant Handler ----

type AIHandler struct {
	aiSvc *service.AIAssistantService
}

func NewAIHandler(aiSvc *service.AIAssistantService) *AIHandler {
	return &AIHandler{aiSvc: aiSvc}
}

// POST /api/v1/ai/chat
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req domain.AIAssistantRequest
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.aiSvc.Chat(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, resp)
}
