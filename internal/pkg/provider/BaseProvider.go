package provider

import (
	"errors"
	"github.com/netandreus/go-forex-rates/internal/pkg/entity"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"time"
)

// BaseProvider implements base provider functionality
type BaseProvider struct {
}

// IsRequestValid validates API call to provider.
func (b *BaseProvider) IsRequestValid(p RatesProvider, ratesRequest model.RatesRequest) (bool, error) {
	baseCurrency := ratesRequest.BaseCurrency
	date := ratesRequest.Date.Format(util.DateFormatEu)
	symbols := ratesRequest.Symbols
	if date != "" {
		dateObject, _ := time.ParseInLocation(util.DateFormatEu, date, p.GetLocation())
		// Check date is not in future
		if dateObject.After(util.GetToday(p.GetLocation())) {
			return false, errors.New("date should not be in future")
		}
	}
	supportedCurrencies := p.GetSupportedCurrencies()

	// Check baseCurrency is supported
	if !util.Contains(supportedCurrencies, baseCurrency) {
		return false, errors.New("base currency " + baseCurrency + " does not supported")
	}

	// Check all symbols are supported
	for _, symbol := range symbols {
		if !util.Contains(supportedCurrencies, symbol) {
			return false, errors.New("quoted currency " + symbol + " does not supported")
		}
	}

	// Check date not before provider start date
	providerStartDateStr := p.GetConfig().HistoricalStartDate
	if providerStartDateStr != "" {
		providerStartDate, _ := time.Parse(util.DateFormatEu, providerStartDateStr)
		if ratesRequest.Date.Before(providerStartDate) {
			return false, errors.New("request date is before provider " + p.GetCode() +
				" historical_start_date: " + providerStartDateStr)
		}
	}
	return true, nil
}

// BuildEntity builds entity with given rates
func (b *BaseProvider) BuildEntity(endpoint string, providerCode string, baseCurrency string, quotedCurrency string, rate float64, rateTime time.Time, providerTime time.Time) *entity.CurrencyRate {
	return &entity.CurrencyRate{
		Provider:              providerCode,
		Endpoint:              endpoint,
		BaseCurrency:          baseCurrency,
		QuotedCurrency:        quotedCurrency,
		RateDate:              rateTime.Format(util.DateFormatEu),
		ProviderGeneratedTime: providerTime.UTC(),
		RequestTime:           time.Now().UTC(),
		Value:                 util.ToFixed(rate, 6),
	}
}

// GetRateGenerationTime returns time, when provider generates history rates for today and we can fetch it.
func (b *BaseProvider) GetRateGenerationTime(timeStr string) time.Time {
	date, _ := time.Parse(util.TimeFormat, timeStr)
	return date
}
