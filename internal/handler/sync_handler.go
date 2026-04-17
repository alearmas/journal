package handler

import (
	"alearmas/tradingJournal/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles the PPI sync endpoint.
type SyncHandler struct {
	svc *service.SyncService
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// SyncResponse summarises the result of a sync operation.
type SyncResponse struct {
	MovimientosImported int      `json:"movimientos_imported" example:"3"`
	MovimientosSkipped  int      `json:"movimientos_skipped"  example:"5"`
	CaucionesImported   int      `json:"cauciones_imported"   example:"2"`
	CaucionesSkipped    int      `json:"cauciones_skipped"    example:"0"`
	Errors              []string `json:"errors,omitempty"`
}

// Sync godoc
// @Summary      Sync from PPI
// @Description  Fetch capital movements and cauciones from PPI for the given date range and import them. Defaults to the last 30 days.
// @Tags         sync
// @Produce      json
// @Param        date_from query string false "Start date YYYY-MM-DD (default: 30 days ago)" example(2026-01-01)
// @Param        date_to   query string false "End date YYYY-MM-DD (default: today)"         example(2026-03-06)
// @Success      200  {object} SyncResponse
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /sync [post]
func (h *SyncHandler) Sync(c *gin.Context) {
	dateTo := time.Now()
	dateFrom := dateTo.AddDate(0, 0, -30)

	if s := c.Query("date_from"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date_from, expected YYYY-MM-DD"})
			return
		}
		dateFrom = t
	}

	if s := c.Query("date_to"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date_to, expected YYYY-MM-DD"})
			return
		}
		dateTo = t
	}

	result, err := h.svc.Sync(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SyncResponse{
		MovimientosImported: result.MovimientosImported,
		MovimientosSkipped:  result.MovimientosSkipped,
		CaucionesImported:   result.CaucionesImported,
		CaucionesSkipped:    result.CaucionesSkipped,
		Errors:              result.Errors,
	})
}
