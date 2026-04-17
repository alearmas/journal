package handler

import (
	"alearmas/tradingJournal/internal/compare"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CompareResponse is the result of comparing caucion vs alternative instruments.
type CompareResponse struct {
	Days       int    `json:"days"        example:"1"`
	Principal  string `json:"principal"   example:"1000000.00"`
	CaucionNet string `json:"caucion_net" example:"1904.00"`
	PFGross    string `json:"pf_gross"    example:"2222.22"`
	MMGross    string `json:"mm_gross"    example:"2083.33"`
}

// Compare godoc
// @Summary      Compare investment instruments
// @Description  Compare caución net return vs plazo fijo and money market gross returns for the same principal and term.
// @Tags         compare
// @Produce      json
// @Param        principal   query string true  "Principal amount (decimal)"    example(1000000.00)
// @Param        days        query int    true  "Term in days"                  example(1)
// @Param        caucion_tna query string true  "Caución TNA (%, e.g. 85.5)"   example(85.5)
// @Param        fees        query string false "Caución fees (decimal)"        example(50.00)
// @Param        taxes       query string false "Caución taxes (decimal)"       example(421.00)
// @Param        pf_tna      query string true  "Plazo fijo TNA (%, e.g. 80.0)" example(80.0)
// @Param        mm_tna      query string true  "Money market TNA (%, e.g. 75)" example(75.0)
// @Success      200  {object} CompareResponse
// @Failure      400  {object} ErrorResponse
// @Router       /compare [get]
func Compare(c *gin.Context) {
	parseDecimal := func(key string) (decimal.Decimal, error) {
		return decimal.NewFromString(c.DefaultQuery(key, "0"))
	}

	principal, err := parseDecimal("principal")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid principal"})
		return
	}

	daysStr := c.DefaultQuery("days", "1")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid days: must be a positive integer"})
		return
	}

	cTna, err := parseDecimal("caucion_tna")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid caucion_tna"})
		return
	}
	fees, err := parseDecimal("fees")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid fees"})
		return
	}
	taxes, err := parseDecimal("taxes")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid taxes"})
		return
	}
	pfTna, err := parseDecimal("pf_tna")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid pf_tna"})
		return
	}
	mmTna, err := parseDecimal("mm_tna")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid mm_tna"})
		return
	}

	out := compare.Run(compare.CompareInput{
		Principal:    principal,
		Days:         days,
		CaucionTNA:   cTna,
		CaucionFees:  fees,
		CaucionTaxes: taxes,
		PFTNA:        pfTna,
		MMTNA:        mmTna,
	})

	c.JSON(http.StatusOK, CompareResponse{
		Days:       days,
		Principal:  principal.StringFixed(2),
		CaucionNet: out.CaucionNet.StringFixed(2),
		PFGross:    out.PFGross.StringFixed(2),
		MMGross:    out.MMGross.StringFixed(2),
	})
}
