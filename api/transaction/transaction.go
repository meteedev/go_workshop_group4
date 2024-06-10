package transaction

import "time"

type TransactionRequest struct {
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	TransactionType string    `json:"transaction_type"`
	Note            string    `json:"note"`
	ImageUrl        string    `json:"image_url"`
	SpenderID       int64     `json:"spender_id"`
}

type TransactionResponse struct {
	ID              int64      `json:"id"`
	Date            *time.Time `json:"date"`
	Amount          float64    `json:"amount"`
	Category        string     `json:"category"`
	TransactionType string     `json:"transaction_type"`
	Note            string     `json:"note"`
	ImageUrl        string     `json:"image_url"`
}
