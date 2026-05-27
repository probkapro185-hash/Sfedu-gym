package httphandler

import (
	"net/http"
	"time"

	"github.com/sfedu-crm/internal/domain"
	"github.com/sfedu-crm/internal/middleware"
	"github.com/sfedu-crm/internal/repository"
	"github.com/sfedu-crm/internal/service"
)

// ---- Finance Handler ----

type FinanceHandler struct {
	financeSvc *service.FinanceService
}

func NewFinanceHandler(financeSvc *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{financeSvc: financeSvc}
}

// POST /api/v1/finance/topup  — пополнить баланс клиента (менеджер/админ)
func (h *FinanceHandler) TopUpBalance(w http.ResponseWriter, r *http.Request) {
	// Получаем ID клиента из тела (менеджер/админ задают явно)
	var body struct {
		ClientID int64   `json:"client_id"`
		Amount   float64 `json:"amount"`
	}
	if err := decode(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.financeSvc.TopUpBalance(r.Context(), body.ClientID, domain.TopUpBalanceInput{Amount: body.Amount})
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, payment)
}

// POST /api/v1/finance/me/topup  — клиент пополняет свой баланс
func (h *FinanceHandler) TopUpMyBalance(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	var input domain.TopUpBalanceInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.financeSvc.TopUpBalance(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, payment)
}

// GET /api/v1/finance/payments  — только ADMIN: все платежи
func (h *FinanceHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.PaymentFilter{}

	if opType := q.Get("operation_type"); opType != "" {
		filter.OperationType = domain.OperationType(opType)
	}
	if svcType := q.Get("service_type"); svcType != "" {
		filter.ServiceType = domain.ServiceType(svcType)
	}
	if from := q.Get("date_from"); from != "" {
		t, _ := time.Parse(time.DateOnly, from)
		filter.DateFrom = &t
	}
	if to := q.Get("date_to"); to != "" {
		t, _ := time.Parse(time.DateOnly, to)
		filter.DateTo = &t
	}

	payments, err := h.financeSvc.ListPayments(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, payments)
}

// GET /api/v1/finance/summary  — только ADMIN: сводка
func (h *FinanceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.financeSvc.GetSummary(r.Context(), domain.PaymentFilter{})
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, summary)
}

// GET /api/v1/finance/me/payments  — клиент: своя история платежей
func (h *FinanceHandler) GetMyPayments(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	payments, err := h.financeSvc.GetClientPayments(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, payments)
}

// ---- Shop Handler ----

type ShopHandler struct {
	shopSvc *service.ShopService
}

func NewShopHandler(shopSvc *service.ShopService) *ShopHandler {
	return &ShopHandler{shopSvc: shopSvc}
}

// GET /api/v1/shop/products?category=
func (h *ShopHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter := repository.ProductFilter{}
	if cat := r.URL.Query().Get("category"); cat != "" {
		c := domain.ProductCategory(cat)
		filter.Category = &c
	}
	active := true
	filter.IsActive = &active

	products, err := h.shopSvc.ListProducts(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, products)
}

// GET /api/v1/shop/products/{id}
func (h *ShopHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	product, err := h.shopSvc.GetProduct(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, product)
}

// POST /api/v1/shop/products  — менеджер/админ
func (h *ShopHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateProductInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.shopSvc.CreateProduct(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, product)
}

// PUT /api/v1/shop/products/{id}
func (h *ShopHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var input domain.CreateProductInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.shopSvc.UpdateProduct(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, product)
}

// DELETE /api/v1/shop/products/{id}  — только админ
func (h *ShopHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.shopSvc.DeleteProduct(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusNoContent, nil)
}

// POST /api/v1/shop/purchase  — клиент покупает товар
func (h *ShopHandler) PurchaseProduct(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	var input domain.PurchaseProductInput
	if err := decode(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := h.shopSvc.PurchaseProduct(r.Context(), userID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusCreated, sub)
}

// GET /api/v1/shop/my-subscriptions  — активные абонементы клиента
func (h *ShopHandler) GetMySubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	subs, err := h.shopSvc.GetClientSubscriptions(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	respond(w, http.StatusOK, subs)
}
