package provider

import (
	"github.com/netandreus/go-forex-rates/internal/pkg/entity"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"time"
)

// RatesProvider for all currency rates providers
type RatesProvider interface {
	GetCode() string
	GetConfig() model.ProviderConfig
	GetHistoricalRates(ratesRequest model.RatesRequest) (model.RatesResponse, error)
	GetLatestRates(ratesRequest model.RatesRequest) (model.RatesResponse, error)
	PreloadRates(date time.Time, save bool) (map[string]float64, map[string]float64, time.Time, error)
	GetRateGenerationTime() time.Time
	GetSupportedCurrencies() []string
	IsRequestValid(ratesRequest model.RatesRequest) (bool, error)
	GetLocation() *time.Location
	BuildEntity(endpoint string, baseCurrency string, quotedCurrency string, rate float64, rateDate time.Time, providerDate time.Time) *entity.CurrencyRate
}
