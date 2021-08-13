package model

import "encoding/json"

// RatesResponse represents result for RatesRequest
type RatesResponse struct {
	// Rates exchange rate data for the currencies you have requested.
	Rates map[string]float64 `json:"rates"`

	// Timestamp when provider generate rates in Rates if it single-pair, or first pair if multiple symbols in Rates
	Timestamp int64 `json:"timestamp"`
}

// String returns string representation of JSON of this key structure
func (r *RatesResponse) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromString fills receiver with actual values from json string in arguments
func (r *RatesResponse) FromString(str string) error {
	return json.Unmarshal([]byte(str), r)
}
