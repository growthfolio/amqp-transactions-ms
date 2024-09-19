package dto

import "time"

type Transaction struct {
	ID           string    `json:"id"`
	Date         time.Time `json:"date"`
	Document     string    `json:"document"`
	Name         string    `json:"name"`
	Age          int       `json:"age"`
	Amount       float64   `json:"amount"`
	Installments int       `json:"installments"`
}
