// Package fixer implements fixer.io provider related code
package fixer

import (
	"encoding/json"
	"errors"
	"github.com/netandreus/go-forex-rates/internal/pkg/entity"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"gorm.io/gorm"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

// Code fixer provider code
const Code = "fixer"

// Provider implements fixer provider structure
type Provider struct {
	provider.BaseProvider
	config model.ProviderConfig
	db     *gorm.DB
	code   string
}

// New constructor
func New(db *gorm.DB, config *model.ApplicationConfig) *Provider {
	// Build provider
	provider := &Provider{
		code:   Code,
		db:     db,
		config: config.Providers[Code],
	}
	return provider
}

// GetHistoricalRates searches for rates in internal DB, fetch from provider API if needed and saves to internal DB
func (p Provider) GetHistoricalRates(serviceRequest model.RatesRequest) (model.RatesResponse, error) {
	var (
		err                   error
		rates                 map[string]float64
		providerGeneratedTime time.Time
		resp                  *http.Response
		body                  []byte
		serviceResponse       model.RatesResponse
	)
	// Validate request
	if _, err := p.IsRequestValid(serviceRequest); err != nil {
		return model.RatesResponse{}, errors.New("request is invalid. " + err.Error())
	}

	// Fetch rates
	apiKey := p.config.APIKey
	symbolsStr := strings.Join(serviceRequest.Symbols, ",")
	dateStr := serviceRequest.Date.Format(util.DateFormatEu)
	url := "https://data.fixer.io/api/" + dateStr + "?access_key=" + apiKey + "&base=" + serviceRequest.BaseCurrency + "&symbols=" + symbolsStr
	if resp, err = http.Get(url); err != nil {
		return serviceResponse, err
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return serviceResponse, err
	}
	if rates, _, providerGeneratedTime, err = p.getRatesFromResponse(body); err != nil {
		return serviceResponse, err
	}

	// Return rates
	serviceResponse = model.RatesResponse{
		Rates:     rates,
		Timestamp: providerGeneratedTime.Unix(),
	}
	return serviceResponse, nil
}

// GetLatestRates return actual rates
func (p Provider) GetLatestRates(serviceRequest model.RatesRequest) (model.RatesResponse, error) {
	var (
		serviceResponse       model.RatesResponse
		err                   error
		rates                 map[string]float64
		providerGeneratedTime time.Time
		resp                  *http.Response
		body                  []byte
	)

	// Validate request
	if _, err = p.IsRequestValid(serviceRequest); err != nil {
		return serviceResponse, errors.New("request is invalid. " + err.Error())
	}

	// Load rates from fixer.io
	apiKey := p.config.APIKey
	symbolsStr := strings.Join(serviceRequest.Symbols, ",")
	url := "https://data.fixer.io/api/latest?access_key=" + apiKey + "&base=" + serviceRequest.BaseCurrency + "&symbols=" + symbolsStr
	if resp, err = http.Get(url); err != nil {
		return serviceResponse, err
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return serviceResponse, err
	}
	if rates, _, providerGeneratedTime, err = p.getRatesFromResponse(body); err != nil {
		return serviceResponse, err
	}
	serviceResponse = model.RatesResponse{
		Rates:     rates,
		Timestamp: providerGeneratedTime.Unix(),
	}
	return serviceResponse, nil
}

// PreloadRates preload all available rates for given date
func (p Provider) PreloadRates(date time.Time, save bool) (map[string]float64, map[string]float64, time.Time, error) {
	return make(map[string]float64), make(map[string]float64), time.Time{}, nil
}

// GetRateGenerationTime returns historical rates generated time on provider side
func (p Provider) GetRateGenerationTime() time.Time {
	return p.BaseProvider.GetRateGenerationTime(p.config.RatesGeneratedTime)
}

// GetCode returns provider code
func (p Provider) GetCode() string {
	return p.code
}

// GetConfig returns provider config
func (p Provider) GetConfig() model.ProviderConfig {
	return p.config
}

// IsRequestValid validates API call to provider.
func (p Provider) IsRequestValid(ratesRequest model.RatesRequest) (bool, error) {
	return p.BaseProvider.IsRequestValid(p, ratesRequest)
}

// BuildEntity builds entity with given rates
func (p Provider) BuildEntity(endpoint string, baseCurrency string, quotedCurrency string, rate float64, rateDate time.Time, providerDate time.Time) *entity.CurrencyRate {
	e := p.BaseProvider.BuildEntity(endpoint, p.GetCode(), baseCurrency, quotedCurrency, rate, rateDate, providerDate)
	return e
}

// GetLocation returns location for current provider
func (p Provider) GetLocation() *time.Location {
	location, _ := time.LoadLocation("UTC")
	return location
}

// GetSupportedCurrencies returns list of currencies, supported by provider
func (p Provider) GetSupportedCurrencies() []string {
	return p.config.SupportedCurrencies
}

// getRatesFromResponse parse response and get fetch rates from it
func (p Provider) getRatesFromResponse(body []byte) (map[string]float64, map[string]float64, time.Time, error) {
	var (
		err                    error
		apiJson                model.SuccessApiResponse
		directRates            = make(map[string]float64)
		reverseRates           = make(map[string]float64)
		normalizedDirectRates  = make(map[string]float64)
		normalizedReverseRates = make(map[string]float64)
		providerGeneratedTime  time.Time
	)
	// Rates
	if err = json.Unmarshal(body, &apiJson); err != nil {
		return directRates, reverseRates, time.Time{}, err
	}

	for cur, directRate := range apiJson.Rates {
		normalizedDirectRates[cur] = math.Round(directRate*1000000) / 1000000
	}

	// Provider generated time
	providerGeneratedTime = time.Unix(apiJson.Timestamp, 0)
	return normalizedDirectRates, normalizedReverseRates, providerGeneratedTime, nil
}
