package model

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"strconv"
	"strings"
	"time"
)

// RatesRequest  is request to internal storage subsystem (multi-level cache)
type RatesRequest struct {
	// Requested endpoint
	Endpoint string `json:"endpoint"`

	// Requested provider code
	ProviderCode string `json:"provider_code"`

	// ProviderLocation name by code and provider config
	ProviderLocationName string `json:"provider_location_name"`

	// Requested currency rates date
	Date time.Time `json:"date"`

	// Requested base currency
	BaseCurrency string `json:"base_currency"`

	// Requested quoted currencies
	Symbols []string `json:"symbols"`

	// If true - do not use any type of cache, makes provider API request this case
	Force bool

	// Is this request build from another
	IsForwarded bool
}

// String returns string representation of JSON of this key structure
func (r *RatesRequest) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromString fills receiver with actual values from json string in arguments
func (r *RatesRequest) FromString(str string) error {
	return json.Unmarshal([]byte(str), r)
}

// FromGinContext fills with data from HTTP Request
func (r *RatesRequest) FromGinContext(c *gin.Context, config *ApplicationConfig, endpoint string) error {
	var (
		err   error
		date  time.Time
		force bool
	)
	// Endpoint check
	if endpoint != util.EndpointLatest && endpoint != util.EndpointHistorical {
		return errors.New("unsupported API endpoint. Allows only(historical, latest). Received: " + endpoint)
	}
	r.Endpoint = endpoint

	// Date check
	if endpoint == util.EndpointHistorical {
		date, err = time.ParseInLocation(util.DateFormatEu, c.Param("date"), time.UTC)
		if err != nil {
			return err
		}
		r.Date = date
	}

	// Force check
	force, _ = strconv.ParseBool(c.Query("force"))
	r.Force = force

	// Base currency check
	baseCurrency := c.Query("base")
	if len(baseCurrency) != 3 {
		return errors.New("unsupported base currency. Received: " + baseCurrency)
	}
	r.BaseCurrency = baseCurrency

	// Provider code check
	providerCode := c.Param("provider")
	r.ProviderCode = providerCode

	// Provider location name
	r.ProviderLocationName = config.Providers[providerCode].Location

	// Symbols check
	symbolsStr := c.Query("symbols")
	symbols := strings.Split(symbolsStr, ",")
	symbols = util.UniqueStringSlice(symbols)
	r.Symbols = symbols
	return nil
}

// IsEqualCurrencyRequest detects such type of request
func (r *RatesRequest) IsEqualCurrencyRequest() bool {
	return len(r.Symbols) == 1 && r.BaseCurrency == r.Symbols[0]
}
