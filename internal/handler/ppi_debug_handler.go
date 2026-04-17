package handler

import (
	"alearmas/tradingJournal/internal/adapter/ppi"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// PPIDebugHandler exposes raw PPI API responses for debugging and field mapping validation.
type PPIDebugHandler struct {
	client        *ppi.Client
	accountNumber string
}

// NewPPIDebugHandler creates a new PPIDebugHandler.
func NewPPIDebugHandler(client *ppi.Client, accountNumber string) *PPIDebugHandler {
	return &PPIDebugHandler{client: client, accountNumber: accountNumber}
}

func (h *PPIDebugHandler) parseDateRange(c *gin.Context) (dateFrom, dateTo time.Time, ok bool) {
	dateTo = time.Now()
	dateFrom = dateTo.AddDate(0, 0, -30)

	if s := c.Query("date_from"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date_from, expected YYYY-MM-DD"})
			return time.Time{}, time.Time{}, false
		}
		dateFrom = t
	}
	if s := c.Query("date_to"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date_to, expected YYYY-MM-DD"})
			return time.Time{}, time.Time{}, false
		}
		dateTo = t
	}
	return dateFrom, dateTo, true
}

// RawMovements godoc
// @Summary      Raw PPI movements
// @Description  Returns the unprocessed movement list from PPI. Use this to inspect field names and values before adjusting sync mappings.
// @Tags         ppi-debug
// @Produce      json
// @Param        date_from query string false "Start date YYYY-MM-DD (default: 30 days ago)" example(2026-01-01)
// @Param        date_to   query string false "End date YYYY-MM-DD (default: today)"         example(2026-03-06)
// @Success      200  {array}  ppi.Movement
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /ppi/movements [get]
func (h *PPIDebugHandler) RawMovements(c *gin.Context) {
	dateFrom, dateTo, ok := h.parseDateRange(c)
	if !ok {
		return
	}

	movements, err := h.client.GetMovements(
		c.Request.Context(),
		h.accountNumber,
		dateFrom.Format(time.RFC3339),
		dateTo.Format(time.RFC3339),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, movements)
}

// RawOrders godoc
// @Summary      Raw PPI orders
// @Description  Returns the unprocessed order list from PPI. Use this to verify caucion field mapping (operation, price=TNA, amount=principal, operationMaxDate=maturity).
// @Tags         ppi-debug
// @Produce      json
// @Param        date_from query string false "Start date YYYY-MM-DD (default: 30 days ago)" example(2026-01-01)
// @Param        date_to   query string false "End date YYYY-MM-DD (default: today)"         example(2026-03-06)
// @Success      200  {array}  ppi.Order
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /ppi/orders [get]
func (h *PPIDebugHandler) RawOrders(c *gin.Context) {
	dateFrom, dateTo, ok := h.parseDateRange(c)
	if !ok {
		return
	}

	orders, err := h.client.GetOrders(
		c.Request.Context(),
		h.accountNumber,
		dateFrom.Format(time.RFC3339),
		dateTo.Format(time.RFC3339),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
