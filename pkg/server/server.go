// Package server provides service engine constructor and related methods
// to work with server.Server structure
package server

import (
	"errors"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/gocolly/colly/v2"
	"github.com/netandreus/go-forex-rates/internal/pkg/controller"
	"github.com/netandreus/go-forex-rates/internal/pkg/logger"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider"
	"github.com/netandreus/go-forex-rates/internal/pkg/service"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Server is server engine instance
type Server struct {
	// DIC container (dig)
	container *dig.Container

	// http-server (gin)
	http *gin.Engine

	// database client
	db *gorm.DB

	// part of application config for engine
	config model.EngineConfig

	// providers registry
	registry *provider.Registry
}

// New factory method to construct new server instance
func New() (*Server, error) {
	var (
		err    error
		engine = &Server{}
		c      *dig.Container
	)

	// Build container
	c, err = engine.buildContainer()
	if err != nil {
		return engine, errors.New("error loading dependency injection container")
	}

	// Init onClose after container is initialized
	c.Invoke(engine.initOnClose)

	// Set http engine
	c.Invoke(func(gin *gin.Engine, config *model.ApplicationConfig, db *gorm.DB, registry *provider.Registry) {
		engine.http = gin
		engine.config = config.Engine
		engine.db = db
		engine.registry = registry
	})

	return engine, nil
}

// Run starts cron listener, first time currency rates preload and REST HTTP server
func (r *Server) Run() error {
	var err error
	// Init auto-refresh currency rates by cron
	if err = r.container.Invoke(r.initAutoRefreshRates); err != nil {
		return err
	}

	// Run once when start
	if err = r.container.Invoke(r.initFirstRefresh); err != nil {
		return err
	}

	// Run http server
	if err = r.http.Run(":" + strconv.Itoa(r.config.Port)); err != nil {
		return err
	}
	return nil
}

// GetListenPort returns REST HTTP server listen port
func (r *Server) GetListenPort() int {
	return r.config.Port
}

// ContainerInvoke invokes passed function with automatic dependency resolution to DIC
func (r *Server) ContainerInvoke(function interface{}) error {
	return r.container.Invoke(function)
}

// buildContainer builds DIC (dependency injection container) and set logger
func (r *Server) buildContainer() (*dig.Container, error) {
	var err error

	r.container = dig.New()

	// Build Services
	if err = r.buildServices(); err != nil {
		panic(err)
	}

	// Build Controllers
	if err = r.buildControllers(); err != nil {
		panic(err)
	}

	// Set logger
	log.SetOutput(new(logger.Logger))
	log.SetFlags(0)

	return r.container, nil
}

// buildControllers builds controllers as a DIC-service
func (r *Server) buildControllers() error {
	var err error
	if r.container == nil {
		return errors.New("container does not initialized before buildControllers() called")
	}
	if err = r.container.Provide(controller.NewApiController); err != nil {
		return err
	}
	return nil
}

// buildServices builds DIC-services
func (r *Server) buildServices() error {
	var err error
	if r.container == nil {
		return errors.New("container does not initialized before buildServices() called")
	}

	// Service *cron.Cron
	if err = r.container.Provide(service.BuildCron); err != nil {
		return err
	}

	// Service *colly.Collector
	if err = r.container.Provide(service.BuildColly); err != nil {
		return err
	}

	// Service: *model.Config
	if err = r.container.Provide(service.BuildConfig); err != nil {
		return err
	}

	// Service: *gorm.DB
	if err = r.container.Provide(service.BuildDatabase); err != nil {
		return err
	}

	// Service: *cache.Cache
	if err = r.container.Provide(service.BuildCache); err != nil {
		return err
	}

	// Service: *gin.Server
	if err = r.container.Provide(service.BuildHttp); err != nil {
		return err
	}

	// Service: *Registry
	if err = r.container.Provide(provider.BuildRegistry); err != nil {
		return err
	}
	return nil
}

// getProvidersNeedToRatesPreload fetch providers need to preloading rates
func (r *Server) getProvidersNeedToRatesPreload(config *model.ApplicationConfig) []provider.RatesProvider {
	var (
		providers []provider.RatesProvider // providers need to refresh
	)
	for providerCode, providerConfig := range config.Providers {
		if providerConfig.HistoricalPreload {
			provider, _ := r.registry.GetProvider(providerCode)
			providers = append(providers, provider)
		}
	}
	return providers
}

// initFirstRefresh preloads historical currency rate for Emirates provider
func (r *Server) initFirstRefresh(coll *colly.Collector, config *model.ApplicationConfig) error {
	var (
		endDate   time.Time
		providers []provider.RatesProvider // providers need to refresh
	)
	// Fetch providers need to
	providers = r.getProvidersNeedToRatesPreload(config)

	if len(providers) == 0 {
		return nil
	}

	for _, provider := range providers {
		rateGenerationTime := provider.GetRateGenerationTime()
		now := time.Now()
		if now.Hour() >= rateGenerationTime.Hour() && now.Minute() > rateGenerationTime.Minute() && now.Second() >= rateGenerationTime.Second() {
			endDate = util.GetToday(time.UTC)
		} else {
			endDate = util.GetYesterday(time.UTC)
		}
		r.refreshCurrencyRates(provider, coll, endDate)
	}
	return nil
}

// initAutoRefreshRates initialize preload rates process now
func (r *Server) initAutoRefreshRates(coll *colly.Collector, cron *gocron.Scheduler, config *model.ApplicationConfig) {
	var (
		today              = util.GetToday(time.UTC)
		providers          []provider.RatesProvider // providers need to refresh
		rateGenerationTime time.Time
	)
	log.Print(color.GreenString("Application started"))

	// Fetch providers need to
	providers = r.getProvidersNeedToRatesPreload(config)

	for _, provider := range providers {
		rateGenerationTime = provider.GetRateGenerationTime()

		// Run Cron
		// 	job, _ := cron.Every(60).Seconds().Do(func() {
		job, _ := cron.Every(1).Day().At(rateGenerationTime.Format(util.TimeFormat)).Do(func() {
			log.Print("Cron event triggered")
			// Options
			//RefreshCurrencyRates(currencyRatesRepository, coll, time.Parse(constants.DateFormatEu, "2018-11-01"))
			r.refreshCurrencyRates(provider, coll, today)
		})
		job.SingletonMode()
	}
	cron.StartAsync()
}

// refreshCurrencyRates preload rates from provider start date to given date for passed provider
func (r *Server) refreshCurrencyRates(
	provider provider.RatesProvider,
	coll *colly.Collector,
	endDate time.Time) {
	// Form array of dates
	startDate, _ := r.getRatesDateStart(provider)
	dateRange := util.GetDateRangeArr(startDate, endDate)
	dateRangeLength := len(dateRange)
	if dateRangeLength == 0 {
		log.Print("Currency rates database is filled")
		return
	}

	// Register colly callbacks
	coll.OnResponse(func(resp *colly.Response) {
		dateStr := resp.Request.URL.Query().Get("date")
		date, _ := time.Parse(util.DateFormatRu, dateStr)
		provider.PreloadRates(date, true)
		log.Print("Fetched for date " + dateStr)
	})

	// Add jobs
	message := "Currency rates database needs filling"
	message += " from " + startDate.Format(util.DateFormatEu)
	message += " to " + endDate.Format(util.DateFormatEu)
	log.Print(color.YellowString(message))
	for i := 0; i < dateRangeLength; i++ {
		date := dateRange[i]
		dateString := date.Format(util.DateFormatRu)
		url := "https://www.centralbank.ae/en/fx-rates-ajax?date=" + dateString + "&v=2"
		coll.Visit(url)
	}
}

// initOnClose creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func (r *Server) initOnClose(cron *gocron.Scheduler) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-c
		log.Println(color.RedString("Receiving stop signal. Exiting..."))
		// Stop cron
		cron.Stop()
		os.Exit(0)
	}()
}

// getRatesDateStart returns given provider's historical rates start date
func (r *Server) getRatesDateStart(provider provider.RatesProvider) (time.Time, error) {
	var count int
	var result time.Time
	row := r.db.Table("currency_rate").
		Select("count(id)").
		Where("endpoint = ?", util.EndpointHistorical).
		Where("provider = ?", provider.GetCode()).
		Row()
	row.Scan(&count)
	if count == 0 {
		// initial start date
		providerStartDateStr := provider.GetConfig().HistoricalStartDate
		return time.ParseInLocation(util.DateFormatEu, providerStartDateStr, provider.GetLocation())
	}
	row = r.db.Table("currency_rate").
		Select("max(rate_date)").
		Where("endpoint = ?", util.EndpointHistorical).
		Where("provider = ?", provider.GetCode()).
		Row()
	row.Scan(&result)
	result = result.AddDate(0, 0, 1)
	return result, nil
}
