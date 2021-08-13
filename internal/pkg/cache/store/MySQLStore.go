// Package cache_store use for custom store for caching historical currency rates
package cache_store

import (
	"encoding/json"
	"github.com/eko/gocache/store"
	"github.com/netandreus/go-forex-rates/internal/pkg/custom_errors"
	"github.com/netandreus/go-forex-rates/internal/pkg/entity"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// MySQLType represents the storage type as a string value
const MySQLType = "mysql"

// MySQLStore used for store immutable (historical) currency rates in L2 cache in database (MySQL)
type MySQLStore struct {
	client  *gorm.DB
	options *store.Options
}

// NewMySQLStore creates a new store to Memcache instance(s)
func NewMySQLStore(client *gorm.DB, options *store.Options) *MySQLStore {
	if options == nil {
		options = &store.Options{}
	}

	return &MySQLStore{
		client:  client,
		options: options,
	}
}

// Get gets value by key
func (store *MySQLStore) Get(key interface{}) (interface{}, error) {
	value, _, err := store.GetWithTTL(key)
	return value, err
}

// GetWithTTL gets value and ttl by key
func (store *MySQLStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	var (
		serviceRequest = model.RatesRequest{}
		ttl            = time.Duration(1) * time.Second
	)

	// Get key struct
	bytes := []byte(key.(string))
	if err := json.Unmarshal(bytes, &serviceRequest); err != nil {
		return nil, 0, err
	}

	// Load Rates by cache key
	serviceResponse, err := store.loadByKey(serviceRequest)
	if err != nil {
		return nil, 0, err
	}
	return serviceResponse, ttl, nil
}

// Set store value in cache
func (store *MySQLStore) Set(key interface{}, value interface{}, options *store.Options) error {
	var (
		err             error
		serviceRequest  = model.RatesRequest{}
		serviceResponse = &model.RatesResponse{}
	)

	// Get key struct
	bytesKey := []byte(key.(string))
	if err := json.Unmarshal(bytesKey, &serviceRequest); err != nil {
		return err
	}

	// Get value struct
	err = serviceResponse.FromString(value.(string))
	if err != nil {
		return err
	}
	if store.canSet(serviceRequest) {
		store.saveByKey(serviceRequest, *serviceResponse)
	}
	return nil
}

// canSet detect ability to store value in L2 cache
func (store *MySQLStore) canSet(serviceRequest model.RatesRequest) bool {
	var (
		location *time.Location
		today    time.Time
	)

	if serviceRequest.Endpoint == util.EndpointLatest {
		return false
	}
	location, _ = time.LoadLocation(serviceRequest.ProviderLocationName)
	today = util.GetToday(location)
	if util.IsDateEquals(serviceRequest.Date, today) || serviceRequest.Date.After(today) {
		return false
	}

	return true
}

// loadByKey uses internally for fetching from database
func (store *MySQLStore) loadByKey(serviceRequest model.RatesRequest) (string, error) {
	var (
		providerGeneratedTime time.Time
		rates                 = make(map[string]float64)
		ratesResponse         = &model.RatesResponse{}
		entities              []entity.CurrencyRate
		symbols               []string
	)

	// Latest data does not stored in database
	if serviceRequest.Endpoint == util.EndpointLatest {
		return "", custom_errors.NewNotFoundError("Value not found in MySQLCache store")
	}

	// Remove base currency from symbols
	for _, symbol := range serviceRequest.Symbols {
		if symbol != serviceRequest.BaseCurrency {
			symbols = append(symbols, symbol)
		}
	}
	// Load Rates from database
	db := store.client
	result := db.Model(&entity.CurrencyRate{}).
		Select([]string{"quoted_currency", "value", "provider_generated_time"}).
		Where("base_currency = ?", serviceRequest.BaseCurrency).
		Where("quoted_currency IN (?)", serviceRequest.Symbols).
		Where("endpoint = ?", util.EndpointHistorical).
		Where("provider = ?", serviceRequest.ProviderCode).
		Where("rate_date = ?", serviceRequest.Date.Format(util.DateFormatEu)).
		Find(&entities)
	if result.Error != nil {
		return "", custom_errors.NewDatabaseError(result.Error.Error())
	}
	symbolsCount := int64(len(symbols))
	if result.RowsAffected < symbolsCount {
		return "", custom_errors.NewNotFoundError("Value not found in MySQLCache store")
	}
	// Provider generated time
	providerGeneratedTime = entities[0].ProviderGeneratedTime

	// Rates
	for _, entity := range entities {
		rates[entity.QuotedCurrency] = entity.Value
	}

	// Add base currency back to symbols
	if len(symbols) != len(serviceRequest.Symbols) {
		rates[serviceRequest.BaseCurrency] = 1
	}
	ratesResponse = &model.RatesResponse{
		Rates:     rates,
		Timestamp: providerGeneratedTime.Unix(),
	}
	resultStr, _ := ratesResponse.String()
	return resultStr, nil
}

// saveByKey uses internally for save to database
func (store *MySQLStore) saveByKey(key model.RatesRequest, value model.RatesResponse) error {
	for quotedCurrency, rate := range value.Rates {
		entity := &entity.CurrencyRate{
			Endpoint:              key.Endpoint,
			BaseCurrency:          key.BaseCurrency, // Base currency
			QuotedCurrency:        quotedCurrency,   // Quoted currency
			RateDate:              key.Date.Format(util.DateFormatEu),
			ProviderGeneratedTime: time.Unix(value.Timestamp, 0),
			RequestTime:           time.Now().UTC(),
			Value:                 util.ToFixed(rate, 6),
			Provider:              key.ProviderCode,
		}
		// OnConflict is need for On duplicate key cause.
		store.client.Clauses(clause.OnConflict{DoNothing: true}).Create(entity)
	}
	return nil
}

// GetType returns type for store
func (store *MySQLStore) GetType() string {
	return MySQLType
}

// Delete does not affected, as MySQLStore is persistent storage
func (store *MySQLStore) Delete(key interface{}) error {
	return nil
}

// Invalidate does not affected, as MySQLStore is persistent storage
func (store *MySQLStore) Invalidate(options store.InvalidateOptions) error {
	return nil
}

// Clear does not affected, as MySQLStore is persistent storage
func (store *MySQLStore) Clear() error {
	return nil
}
