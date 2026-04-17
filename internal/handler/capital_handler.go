package handler

import (
	"alearmas/tradingJournal/internal/capital"
	"alearmas/tradingJournal/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CapitalHandler handles HTTP requests for capital/balance queries.
type CapitalHandler struct {
	caucionSvc    *service.CaucionService
	movimientoSvc *service.MovimientoService
}

// NewCapitalHandler creates a new CapitalHandler.
func NewCapitalHandler(caucionSvc *service.CaucionService, movimientoSvc *service.MovimientoService) *CapitalHandler {
	return &CapitalHandler{caucionSvc: caucionSvc, movimientoSvc: movimientoSvc}
}

// BrokerSummaryResponse is the capital and P&L summary for a single broker.
type BrokerSummaryResponse struct {
	Broker            string `json:"broker"             example:"Balanz"`
	TotalDeposited    string `json:"total_deposited"    example:"5000000.00"`
	TotalWithdrawn    string `json:"total_withdrawn"    example:"1000000.00"`
	NetBalance        string `json:"net_balance"        example:"4000000.00"`
	DeployedPrincipal string `json:"deployed_principal" example:"3000000.00"`
	Available         string `json:"available"          example:"1000000.00"`
	TotalNetInterest  string `json:"total_net_interest" example:"9520.00"`
	PnLPercent        string `json:"pnl_percent"        example:"0.1904"`
}

// Balance godoc
// @Summary      Capital balance
// @Description  Get capital summary per broker: deposits, withdrawals, deployed principal, available cash, and P&L.
// @Tags         capital
// @Produce      json
// @Param        broker query string false "Filter by broker name" example(Balanz)
// @Success      200  {array}  BrokerSummaryResponse
// @Failure      500  {object} ErrorResponse
// @Router       /capital/balance [get]
func (h *CapitalHandler) Balance(c *gin.Context) {
	brokerFilter := c.Query("broker")

	cauciones, err := h.caucionSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "error listing cauciones: " + err.Error()})
		return
	}

	movimientos, err := h.movimientoSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "error listing movimientos: " + err.Error()})
		return
	}

	summaries := capital.Summarize(movimientos, cauciones, time.Now())

	resp := make([]BrokerSummaryResponse, 0, len(summaries))
	for _, s := range summaries {
		if brokerFilter != "" && s.Broker != brokerFilter {
			continue
		}
		resp = append(resp, BrokerSummaryResponse{
			Broker:            s.Broker,
			TotalDeposited:    s.TotalDeposited.StringFixed(2),
			TotalWithdrawn:    s.TotalWithdrawn.StringFixed(2),
			NetBalance:        s.NetBalance.StringFixed(2),
			DeployedPrincipal: s.DeployedPrincipal.StringFixed(2),
			Available:         s.Available.StringFixed(2),
			TotalNetInterest:  s.TotalNetInterest.StringFixed(2),
			PnLPercent:        s.PnLPercent.StringFixed(4),
		})
	}

	c.JSON(http.StatusOK, resp)
}
