package ppi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Order represents a single order from the PPI API.
// Endpoint: GET /api/1.0/Order/Orders
type Order struct {
	ID              int       `json:"id"`
	InstrumentType  string    `json:"instrumentType"`
	Operation       string    `json:"operation"`   // e.g. "COLOCAR-CAUCIÓN"
	Ticker          string    `json:"ticker"`
	Status          string    `json:"status"`      // e.g. "TERMINADA", "PENDIENTE"
	Date            time.Time `json:"date"`        // trade date
	Settlement      string    `json:"settlement"`  // e.g. "A-24HS"
	Quantity        float64   `json:"quantity"`
	OrderType       string    `json:"orderType"`
	OperationType   string    `json:"operationType"`
	OperationMaxDate time.Time `json:"operationMaxDate"` // maturity date for cauciones
	Price           float64   `json:"price"`            // TNA (%) for cauciones
	Currency        string    `json:"currency"`
	Amount          float64   `json:"amount"`           // principal
	ExternalID      string    `json:"externalID"`
}

// IsCaucion returns true if the order is a caución placement.
func (o Order) IsCaucion() bool {
	return o.Operation == "COLOCAR-CAUCIÓN" || o.Operation == "COLOCAR-CAUCION"
}

// IsSettled returns true if the order has been executed/settled.
func (o Order) IsSettled() bool {
	return o.Status == "TERMINADA" || o.Status == "EJECUTADA" || o.Status == "LIQUIDADA"
}

// GetOrders fetches orders for the given date range.
// dateFrom and dateTo are optional; pass empty string to omit.
func (c *Client) GetOrders(ctx context.Context, accountNumber, dateFrom, dateTo string) ([]Order, error) {
	params := map[string]string{
		"accountNumber": accountNumber,
	}
	if dateFrom != "" {
		params["dateFrom"] = dateFrom
	}
	if dateTo != "" {
		params["dateTo"] = dateTo
	}

	resp, err := c.doGet(ctx, "/api/1.0/Order/Orders", params)
	if err != nil {
		return nil, fmt.Errorf("ppi get orders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ppi get orders: unexpected status %d", resp.StatusCode)
	}

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("ppi get orders: decode: %w", err)
	}

	return orders, nil
}
