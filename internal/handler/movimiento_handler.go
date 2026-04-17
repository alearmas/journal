package handler

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MovimientoHandler handles HTTP requests for capital movements.
type MovimientoHandler struct {
	svc *service.MovimientoService
}

// NewMovimientoHandler creates a new MovimientoHandler.
func NewMovimientoHandler(svc *service.MovimientoService) *MovimientoHandler {
	return &MovimientoHandler{svc: svc}
}

// RegisterMovimientoRequest is the request body for registering a capital movement.
type RegisterMovimientoRequest struct {
	Broker string `json:"broker" example:"Balanz"`
	Date   string `json:"date"   example:"2026-01-01"`
	Amount string `json:"amount" example:"1000000.00"`
	Notes  string `json:"notes"  example:"transferencia inicial"`
}

// MovimientoResponse is the response body for a capital movement.
type MovimientoResponse struct {
	ID        string `json:"id"         example:"a1b2c3d4e5f6240910ab"`
	Broker    string `json:"broker"     example:"Balanz"`
	Date      string `json:"date"       example:"2026-01-01"`
	Type      string `json:"type"       example:"deposito"`
	Amount    string `json:"amount"     example:"1000000.00"`
	Notes     string `json:"notes"      example:"transferencia inicial"`
	CreatedAt string `json:"created_at" example:"2026-01-01T12:00:00Z"`
}

func movimientoToResponse(m domain.Movimiento) MovimientoResponse {
	return MovimientoResponse{
		ID:        m.ID,
		Broker:    m.Broker,
		Date:      m.Date.Format("2006-01-02"),
		Type:      string(m.Type),
		Amount:    m.Amount.StringFixed(2),
		Notes:     m.Notes,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

func (h *MovimientoHandler) registerMovimiento(c *gin.Context, tipo domain.MovimientoTipo) {
	var req RegisterMovimientoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid amount: expected decimal (e.g. 1000000.00)"})
		return
	}

	var date time.Time
	if req.Date != "" {
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date: expected YYYY-MM-DD"})
			return
		}
	}

	m, err := h.svc.Register(c.Request.Context(), service.RegisterMovimientoInput{
		Broker: req.Broker,
		Date:   date,
		Type:   tipo,
		Amount: amount,
		Notes:  req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, movimientoToResponse(m))
}

// Deposit godoc
// @Summary      Register a deposit
// @Description  Register a capital deposit for a broker. Amount must be > 0.
// @Tags         movimientos
// @Accept       json
// @Produce      json
// @Param        body body RegisterMovimientoRequest true "Deposit data"
// @Success      201  {object} MovimientoResponse
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /movimientos/deposit [post]
func (h *MovimientoHandler) Deposit(c *gin.Context) {
	h.registerMovimiento(c, domain.Deposito)
}

// Withdraw godoc
// @Summary      Register a withdrawal
// @Description  Register a capital withdrawal for a broker. Amount must be > 0.
// @Tags         movimientos
// @Accept       json
// @Produce      json
// @Param        body body RegisterMovimientoRequest true "Withdrawal data"
// @Success      201  {object} MovimientoResponse
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /movimientos/withdraw [post]
func (h *MovimientoHandler) Withdraw(c *gin.Context) {
	h.registerMovimiento(c, domain.Retiro)
}

// List godoc
// @Summary      List movimientos
// @Description  Get all capital movements (deposits and withdrawals) ordered by date.
// @Tags         movimientos
// @Produce      json
// @Success      200  {array}  MovimientoResponse
// @Failure      500  {object} ErrorResponse
// @Router       /movimientos [get]
func (h *MovimientoHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]MovimientoResponse, len(items))
	for i, item := range items {
		resp[i] = movimientoToResponse(item)
	}

	c.JSON(http.StatusOK, resp)
}
