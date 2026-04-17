package ppi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Movement represents a single account movement from the PPI API.
// Endpoint: GET /api/1.0/Account/Movements
type Movement struct {
	AgreementDate  time.Time `json:"agreementDate"`
	SettlementDate time.Time `json:"settlementDate"`
	Currency       string    `json:"currency"`
	Amount         float64   `json:"amount"`
	Price          float64   `json:"price"`
	Description    string    `json:"description"`
	Ticker         string    `json:"ticker"`
	Quantity       float64   `json:"quantity"`
	Balance        float64   `json:"balance"`
}

// GetMovements fetches account movements for the given date range.
// dateFrom and dateTo are optional; pass empty string to omit.
func (c *Client) GetMovements(ctx context.Context, accountNumber, dateFrom, dateTo string) ([]Movement, error) {
	params := map[string]string{
		"accountNumber": accountNumber,
	}
	if dateFrom != "" {
		params["dateFrom"] = dateFrom
	}
	if dateTo != "" {
		params["dateTo"] = dateTo
	}

	resp, err := c.doGet(ctx, "/api/1.0/Account/Movements", params)
	if err != nil {
		return nil, fmt.Errorf("ppi get movements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ppi get movements: unexpected status %d", resp.StatusCode)
	}

	var movements []Movement
	if err := json.NewDecoder(resp.Body).Decode(&movements); err != nil {
		return nil, fmt.Errorf("ppi get movements: decode: %w", err)
	}

	return movements, nil
}
