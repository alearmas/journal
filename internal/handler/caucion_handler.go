package handler

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/export"
	"alearmas/tradingJournal/internal/report"
	"alearmas/tradingJournal/internal/service"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CaucionHandler handles HTTP requests for cauciones.
type CaucionHandler struct {
	svc *service.CaucionService
}

// NewCaucionHandler creates a new CaucionHandler.
func NewCaucionHandler(svc *service.CaucionService) *CaucionHandler {
	return &CaucionHandler{svc: svc}
}

// CreateCaucionRequest is the request body for creating a caucion.
type CreateCaucionRequest struct {
	Broker    string `json:"broker"     example:"Balanz"`
	TradeDate string `json:"trade_date" example:"2026-01-10"`
	TermDays  int    `json:"term_days"  example:"1"`
	Principal string `json:"principal"  example:"1000000.00"`
	TNA       string `json:"tna"        example:"85.5"`
	Fees      string `json:"fees"       example:"50.00"`
	Taxes     string `json:"taxes"      example:"421.00"`
	Notes     string `json:"notes"      example:"overnight"`
}

// CaucionResponse is the response body for a caucion.
type CaucionResponse struct {
	ID            string `json:"id"             example:"a1b2c3d4e5f6240910ab"`
	Broker        string `json:"broker"         example:"Balanz"`
	TradeDate     string `json:"trade_date"     example:"2026-01-10"`
	MaturityDate  string `json:"maturity_date"  example:"2026-01-11"`
	TermDays      int    `json:"term_days"      example:"1"`
	Principal     string `json:"principal"      example:"1000000.00"`
	TNA           string `json:"tna"            example:"85.5"`
	GrossInterest string `json:"gross_interest" example:"2375.00"`
	Fees          string `json:"fees"           example:"50.00"`
	Taxes         string `json:"taxes"          example:"421.00"`
	NetInterest   string `json:"net_interest"   example:"1904.00"`
	Notes         string `json:"notes"          example:"overnight"`
	CreatedAt     string `json:"created_at"     example:"2026-01-10T12:00:00Z"`
}

// SummaryResponse is the response for the summary endpoint.
type SummaryResponse struct {
	Count          int    `json:"count"           example:"5"`
	TotalPrincipal string `json:"total_principal" example:"5000000.00"`
	TotalGross     string `json:"total_gross"     example:"11875.00"`
	TotalFees      string `json:"total_fees"      example:"250.00"`
	TotalTaxes     string `json:"total_taxes"     example:"2105.00"`
	TotalNet       string `json:"total_net"       example:"9520.00"`
}

// MonthSummaryResponse is the response for the monthly report endpoint.
type MonthSummaryResponse struct {
	Month          string `json:"month"           example:"2026-01"`
	Count          int    `json:"count"           example:"5"`
	TotalPrincipal string `json:"total_principal" example:"5000000.00"`
	TotalGross     string `json:"total_gross"     example:"11875.00"`
	TotalFees      string `json:"total_fees"      example:"250.00"`
	TotalTaxes     string `json:"total_taxes"     example:"2105.00"`
	TotalNet       string `json:"total_net"       example:"9520.00"`
	WeightedAvgTNA string `json:"weighted_avg_tna" example:"85.5"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

func caucionToResponse(c domain.Caucion) CaucionResponse {
	return CaucionResponse{
		ID:            c.ID,
		Broker:        c.Broker,
		TradeDate:     c.TradeDate.Format("2006-01-02"),
		MaturityDate:  c.MaturityDate.Format("2006-01-02"),
		TermDays:      c.TermDays,
		Principal:     c.Principal.StringFixed(2),
		TNA:           c.TNA.String(),
		GrossInterest: c.GrossInterest.StringFixed(2),
		Fees:          c.Fees.StringFixed(2),
		Taxes:         c.Taxes.StringFixed(2),
		NetInterest:   c.NetInterest.StringFixed(2),
		Notes:         c.Notes,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
	}
}

// Create godoc
// @Summary      Create a caucion
// @Description  Register a new caucion (fixed-term investment). Gross interest = principal × (tna/100) × (term_days/360).
// @Tags         cauciones
// @Accept       json
// @Produce      json
// @Param        body body CreateCaucionRequest true "Caucion data"
// @Success      201  {object} CaucionResponse
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /cauciones [post]
func (h *CaucionHandler) Create(c *gin.Context) {
	var req CreateCaucionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	principal, err := decimal.NewFromString(req.Principal)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid principal: expected decimal (e.g. 1000000.00)"})
		return
	}
	tna, err := decimal.NewFromString(req.TNA)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tna: expected decimal (e.g. 85.5)"})
		return
	}
	fees, err := decimal.NewFromString(req.Fees)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid fees: expected decimal"})
		return
	}
	taxes, err := decimal.NewFromString(req.Taxes)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid taxes: expected decimal"})
		return
	}

	var tradeDate time.Time
	if req.TradeDate != "" {
		tradeDate, err = time.Parse("2006-01-02", req.TradeDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid trade_date: expected YYYY-MM-DD"})
			return
		}
	}

	caucion, err := h.svc.Create(c.Request.Context(), service.CreateCaucionInput{
		Broker:    req.Broker,
		TradeDate: tradeDate,
		TermDays:  req.TermDays,
		Principal: principal,
		TNA:       tna,
		Fees:      fees,
		Taxes:     taxes,
		Notes:     req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, caucionToResponse(caucion))
}

// List godoc
// @Summary      List cauciones
// @Description  Get all cauciones ordered by trade date ascending.
// @Tags         cauciones
// @Produce      json
// @Success      200  {array}  CaucionResponse
// @Failure      500  {object} ErrorResponse
// @Router       /cauciones [get]
func (h *CaucionHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]CaucionResponse, len(items))
	for i, item := range items {
		resp[i] = caucionToResponse(item)
	}

	c.JSON(http.StatusOK, resp)
}

// Summary godoc
// @Summary      Global summary
// @Description  Get aggregate totals (principal, gross, fees, taxes, net) across all cauciones.
// @Tags         cauciones
// @Produce      json
// @Success      200  {object} SummaryResponse
// @Failure      500  {object} ErrorResponse
// @Router       /cauciones/summary [get]
func (h *CaucionHandler) Summary(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var totalPrincipal, totalGross, totalFees, totalTaxes, totalNet decimal.Decimal
	for _, item := range items {
		totalPrincipal = totalPrincipal.Add(item.Principal)
		totalGross = totalGross.Add(item.GrossInterest)
		totalFees = totalFees.Add(item.Fees)
		totalTaxes = totalTaxes.Add(item.Taxes)
		totalNet = totalNet.Add(item.NetInterest)
	}

	c.JSON(http.StatusOK, SummaryResponse{
		Count:          len(items),
		TotalPrincipal: totalPrincipal.StringFixed(2),
		TotalGross:     totalGross.StringFixed(2),
		TotalFees:      totalFees.StringFixed(2),
		TotalTaxes:     totalTaxes.StringFixed(2),
		TotalNet:       totalNet.StringFixed(2),
	})
}

// MonthlyReport godoc
// @Summary      Monthly report
// @Description  Get aggregates and weighted-average TNA for a specific month (YYYY-MM).
// @Tags         cauciones
// @Produce      json
// @Param        month query string true "Month in YYYY-MM format" example(2026-01)
// @Success      200  {object} MonthSummaryResponse
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /cauciones/report [get]
func (h *CaucionHandler) MonthlyReport(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "month query param is required (YYYY-MM)"})
		return
	}

	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	filtered, err := report.FilterByMonth(items, month)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	s := report.SummarizeMonth(filtered, month)

	c.JSON(http.StatusOK, MonthSummaryResponse{
		Month:          s.Month,
		Count:          s.Count,
		TotalPrincipal: s.TotalPrincipal.StringFixed(2),
		TotalGross:     s.TotalGross.StringFixed(2),
		TotalFees:      s.TotalFees.StringFixed(2),
		TotalTaxes:     s.TotalTaxes.StringFixed(2),
		TotalNet:       s.TotalNet.StringFixed(2),
		WeightedAvgTNA: s.WeightedAvgTNA.String(),
	})
}

// ExportCSV godoc
// @Summary      Export CSV
// @Description  Download cauciones as a CSV file. Optionally filter by month.
// @Tags         cauciones
// @Produce      text/csv
// @Param        month query string false "Optional month filter YYYY-MM" example(2026-01)
// @Success      200  {file}   string "cauciones.csv"
// @Failure      400  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /cauciones/export [get]
func (h *CaucionHandler) ExportCSV(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	month := c.Query("month")
	if month != "" {
		filtered, err := report.FilterByMonth(items, month)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		items = filtered
	}

	tmpFile, err := os.CreateTemp("", "cauciones-*.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create temp file"})
		return
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := export.WriteCaucionesCSV(tmpPath, items); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to write CSV"})
		return
	}

	c.FileAttachment(tmpPath, "cauciones.csv")
}
