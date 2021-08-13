package entity

import "time"

// CurrencyRate represents one currency pair exchange rate
type CurrencyRate struct {
	// Id
	ID uint

	// Base currency of currency pair
	BaseCurrency string

	// Quoted currency of currency pair
	QuotedCurrency string

	// Rate value
	Value float64

	// The date on which the rates are requested. String is used so that there is no conversion to the server timezone
	RateDate string `json:"rate_date"`

	// Provider API request time (UTC)
	RequestTime time.Time `json:"request_time"`

	// Provider generated rates time (UTC)
	ProviderGeneratedTime time.Time `json:"provider_generated_time"`

	// Provider of this currency exchange rate
	Provider string

	// Currency rate endpoint
	Endpoint string
}

// TableName returns MySQL table name
func (r CurrencyRate) TableName() string {
	return "currency_rate"
}
