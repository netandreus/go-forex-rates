package model

import (
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"time"
)

// SuccessApiResponse represents success API response
type SuccessApiResponse struct {
	// Success true or false depending on whether or not your API request has succeeded.
	Success bool `json:"success"`

	// Historical true if a request for historical exchange rates was made.
	Historical bool `json:"historical"`

	// Date date for which historical rates were requested.
	Date string `json:"date"`

	// Timestamp the exact date and time (UNIX time stamp) the given rates were collected.
	Timestamp int64 `json:"timestamp"`

	// Base the three-letter currency code of the base currency used for this request.
	Base string `json:"base"`

	// Rates exchange rate data for the currencies you have requested.
	Rates map[string]float64 `json:"rates"`
}

// NewSuccessApiResponse constructor
func NewSuccessApiResponse(serviceRequest RatesRequest, serviceResponse RatesResponse) *SuccessApiResponse {
	return &SuccessApiResponse{
		Success:    true,
		Historical: serviceRequest.Endpoint == util.EndpointHistorical,
		Date:       serviceRequest.Date.Format(util.DateFormatEu),
		Timestamp:  serviceResponse.Timestamp,
		Base:       serviceRequest.BaseCurrency,
		Rates:      serviceResponse.Rates,
	}
}

// NewSuccessApiResponseCurrencyEquals returns response if base currency = quoted currency
func NewSuccessApiResponseCurrencyEquals(serviceRequest RatesRequest) *SuccessApiResponse {
	return &SuccessApiResponse{
		Success:    true,
		Historical: serviceRequest.Endpoint == util.EndpointHistorical,
		Date:       serviceRequest.Date.Format(util.DateFormatEu),
		Timestamp:  time.Now().Unix(),
		Base:       serviceRequest.BaseCurrency,
		Rates:      map[string]float64{serviceRequest.BaseCurrency: 1},
	}
}
