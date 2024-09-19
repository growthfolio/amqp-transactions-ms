package dto

import "time"

type Transaction struct {
	TransactionID string    `json:"transactionId"`
	Date          time.Time `json:"date"`
	ClientID      string    `json:"clientId"`
	Name          string    `json:"name"`
	Age           int       `json:"age"`
	Amount        float64   `json:"amount"`
	Installments  int       `json:"installments"`
}
