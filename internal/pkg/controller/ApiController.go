// Package controller contains all application controllers
package controller

import (
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
	"github.com/gin-gonic/gin"
	"github.com/netandreus/go-forex-rates/internal/pkg/logger"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider/emirates"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"gorm.io/gorm"
	"strings"
	"time"
)

// ApiController is main API controller of application
type ApiController struct {
	db       *gorm.DB
	config   *model.ApplicationConfig
	cache    *cache.ChainCache
	registry *provider.Registry
}

// NewApiController is the constructor
func NewApiController(db *gorm.DB,
	config *model.ApplicationConfig,
	cache *cache.ChainCache,
	registry *provider.Registry) *ApiController {
	return &ApiController{
		db:       db,
		config:   config,
		cache:    cache,
		registry: registry,
	}
}

// isDebug returns bool value is debug mode on?
func (controller *ApiController) isDebug() bool {
	return controller.config.Engine.Mode == gin.DebugMode
}

// Status godoc
// @Summary Using for microservice health-check by Docker
// @Produce json
// @Success 200 {object} model.PingApiResponse
// @Router /status [get]
func (controller *ApiController) Status() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Success"})
	}
	return gin.HandlerFunc(fn)
}

// Historical godoc
// @Summary Get historical currency rates
// @Produce json
// @Param provider path string false "Provider" Enums(emirates, fixer)
// @Param date path string true "Rates date (format YYYY-MM-DD)"
// @Param base query string true "Base currency"
// @Param symbols query string true "Quoted currencies, comme separated"
// @Param force query boolean false "Force do not use any cache"
// @Success 200 {object} model.SuccessApiResponse
// @Router /historical/{provider}/{date} [get]
func (controller *ApiController) Historical() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var (
			err             error
			prov            provider.RatesProvider
			serviceRequest  = model.RatesRequest{}
			serviceResponse = model.RatesResponse{
				Timestamp: 0,
				Rates:     make(map[string]float64),
			}
		)

		// Parse HTTP request params
		if err = serviceRequest.FromGinContext(c, controller.config, util.EndpointHistorical); err != nil {
			c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
			return
		}

		// BaseCurrency = QuotedCurrency ?
		if serviceRequest.IsEqualCurrencyRequest() {
			c.JSON(200, model.NewSuccessApiResponseCurrencyEquals(serviceRequest))
			return
		}

		// Init provider
		if prov, err = controller.registry.GetProvider(serviceRequest.ProviderCode); err != nil {
			c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
			return
		}

		// Result for request today's historical rates
		providerLocation := prov.GetLocation()
		if util.IsDateEquals(serviceRequest.Date, util.GetToday(providerLocation)) {
			serviceRequest.Date = util.GetYesterday(providerLocation)
		}

		// Cache get
		cacheKey, _ := serviceRequest.String()
		cacheValue, err := controller.cache.Get(cacheKey)

		if err != nil && !strings.Contains(err.Error(), "Value not found") {
			c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
			return
		}
		if cacheValue != nil && !serviceRequest.Force {
			if controller.isDebug() {
				logger.LogSuccess("Found", "CACHE")
			}
			// Unmarshall
			if err := serviceResponse.FromString(cacheValue.(string)); err != nil {
				c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
				return
			}
		} else {
			if controller.isDebug() {
				logger.LogWarning("Not found", "CACHE")
			}
			// Get rates
			if controller.isDebug() {
				logger.LogWarning("Request provider \""+serviceRequest.ProviderCode+"\" API", "API")
			}
			if serviceResponse, err = prov.GetHistoricalRates(serviceRequest); err != nil {
				c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
				return
			}
			// Cache set
			if !serviceRequest.Force {
				expiration := time.Duration(controller.config.L1Cache.DefaultExpiration) * time.Second

				// Marshall
				cacheValueStr, err := serviceResponse.String()
				if err != nil {
					c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
					return
				}
				// Set to cache
				controller.cache.Set(cacheKey, cacheValueStr, &store.Options{Expiration: expiration})
				if controller.isDebug() {
					logger.LogSuccess("Saved to cache with key: "+string(cacheKey), "CACHE")
				}
			}
		}

		// Return response
		c.JSON(200, model.NewSuccessApiResponse(serviceRequest, serviceResponse))
	}
	return gin.HandlerFunc(fn)
}

// Latest godoc
// @Summary Get latest currency rates
// @Produce json
// @Param provider path string false "Provider" Enums(emirates, fixer)
// @Param base query string true "Base currency"
// @Param symbols query string true "Quoted currencies, comme separated"
// @Param force query boolean false "Force do not use any cache (except emirates-latest combination)"
// @Success 200 {object} model.SuccessApiResponse
// @Router /latest/{provider} [get]
func (controller *ApiController) Latest() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var (
			err             error
			prov            provider.RatesProvider
			serviceRequest  = model.RatesRequest{}
			serviceResponse = model.RatesResponse{
				Timestamp: 0,
				Rates:     make(map[string]float64),
			}
			date time.Time
			now  = time.Now()
		)

		// Parse HTTP request params
		if err = serviceRequest.FromGinContext(c, controller.config, util.EndpointLatest); err != nil {
			c.JSON(400, model.NewFailedApiResponse(400, "error parsing request. "+err.Error()))
			return
		}

		// Correct service request (Define correct date)
		if serviceRequest.ProviderCode == emirates.Code {
			prov, _ := controller.registry.GetProvider(emirates.Code)
			rateGenerationTime := prov.GetRateGenerationTime()
			if now.Hour() >= rateGenerationTime.Hour() && now.Minute() > rateGenerationTime.Minute() && now.Second() > rateGenerationTime.Second() {
				date = util.GetToday(time.UTC)
			} else {
				date = util.GetYesterday(time.UTC)
				serviceRequest.Endpoint = util.EndpointHistorical
			}
		} else {
			// For L1 cache (in-memory) can save and load
			// date = time.Time{}
			date = util.GetToday(time.UTC)
		}
		serviceRequest.Date = date

		// BaseCurrency = QuotedCurrency ?
		if serviceRequest.IsEqualCurrencyRequest() {
			c.JSON(200, model.NewSuccessApiResponseCurrencyEquals(serviceRequest))
			return
		}

		// Cache get
		cacheKey, _ := serviceRequest.String()
		cacheValue, err := controller.cache.Get(cacheKey)

		if err != nil && !strings.Contains(err.Error(), "Value not found") {
			c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
			return
		}

		if cacheValue != nil && !serviceRequest.Force {
			if controller.isDebug() {
				logger.LogSuccess("Found", "CACHE")
			}
			// Unmarshall
			if err := serviceResponse.FromString(cacheValue.(string)); err != nil {
				c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
				return
			}
		} else {
			if controller.isDebug() {
				logger.LogWarning("Not found", "CACHE")
			}
			if prov, err = controller.registry.GetProvider(serviceRequest.ProviderCode); err != nil {
				c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
				return
			}
			// Get rates
			if serviceResponse, err = prov.GetLatestRates(serviceRequest); err != nil {
				c.JSON(400, model.NewFailedApiResponse(400, err.Error()))
				return
			}

			// Cache set
			if !serviceRequest.Force {
				expiration := time.Duration(controller.config.L1Cache.DefaultExpiration) * time.Second
				// Marshall
				cacheValueStr, err := serviceResponse.String()
				if err != nil {
					response := model.NewFailedApiResponse(400, err.Error())
					c.JSON(400, response)
					return
				}
				controller.cache.Set(cacheKey, cacheValueStr, &store.Options{Expiration: expiration})
				if controller.isDebug() {
					logger.LogSuccess("Set cache value with key "+cacheKey, "CACHE")
				}
			}
		}

		// Return response
		c.JSON(200, model.NewSuccessApiResponse(serviceRequest, serviceResponse))
	}
	return gin.HandlerFunc(fn)
}
