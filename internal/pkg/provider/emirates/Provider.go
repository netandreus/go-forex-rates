// Package emirates implements emirates provider related code
package emirates

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/netandreus/go-forex-rates/internal/pkg/custom_errors"
	"github.com/netandreus/go-forex-rates/internal/pkg/entity"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider"
	"github.com/netandreus/go-forex-rates/internal/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Code emirates provider code
const Code = "emirates"

type ApiResponse struct {
	// HTML table in json field
	Table string `json:"table"`

	// Currency rates last updated time
	LastUpdated string `json:"last_updated"`
}

// Provider implements emirates provider structure
type Provider struct {
	provider.BaseProvider
	code   string
	db     *gorm.DB
	config model.ProviderConfig
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

// GetCode returns provider code
func (p Provider) GetCode() string {
	return p.code
}

// GetConfig returns provider config
func (p Provider) GetConfig() model.ProviderConfig {
	return p.config
}

// GetHistoricalRates searches for rates in internal DB, fetch from provider API if needed and saves to internal DB
func (p Provider) GetHistoricalRates(serviceRequest model.RatesRequest) (model.RatesResponse, error) {
	var (
		directRates, reverseRates map[string]float64
		providerGeneratedTime     time.Time
		err                       error
	)

	// Get params
	baseCurrency := serviceRequest.BaseCurrency
	symbols := serviceRequest.Symbols
	date := serviceRequest.Date.Format(util.DateFormatEu)
	force := serviceRequest.Force

	// Validate request
	_, err = p.IsRequestValid(serviceRequest)
	if err != nil {
		return model.RatesResponse{}, errors.New("request is invalid. " + err.Error())
	}

	// Fetch (all rates for date) and save if not
	today := util.GetToday(p.GetLocation())
	dateObject, _ := time.ParseInLocation(util.DateFormatEu, date, p.GetLocation())
	save := !util.IsDateEquals(dateObject, today) && !dateObject.After(today) && !force && !serviceRequest.IsForwarded
	directRates, reverseRates, providerGeneratedTime, err = p.PreloadRates(dateObject, save)

	// Filter by symbols
	serviceResponse := model.RatesResponse{}
	if baseCurrency == "AED" {
		serviceResponse.Rates = p.filterRates(directRates, baseCurrency, symbols)
		serviceResponse.Timestamp = providerGeneratedTime.Unix()
		return serviceResponse, nil
	} else if len(symbols) == 1 && symbols[0] == "AED" {
		serviceResponse.Rates = map[string]float64{symbols[0]: reverseRates[baseCurrency]}
		serviceResponse.Timestamp = providerGeneratedTime.Unix()
		return serviceResponse, nil
	} else {
		serviceResponse.Rates = make(map[string]float64)
		serviceResponse.Timestamp = providerGeneratedTime.Unix()
		return serviceResponse, custom_errors.NewBadRequestError("base currency should be AED, or symbols should be [AED]")
	}
}

// GetLatestRates provides yesterday-defined rate for today latest rate if time < 23:00
func (p Provider) GetLatestRates(serviceRequest model.RatesRequest) (model.RatesResponse, error) {
	var now = time.Now()
	rateGenerationTime := p.GetRateGenerationTime()
	if now.Hour() >= rateGenerationTime.Hour() && now.Minute() > rateGenerationTime.Minute() && now.Second() > rateGenerationTime.Second() {
		serviceRequest.Date = util.GetToday(time.UTC)
	} else {
		serviceRequest.Date = util.GetYesterday(time.UTC)
	}
	serviceRequest.Force = false
	serviceRequest.IsForwarded = true
	return p.GetHistoricalRates(serviceRequest)
}

// PreloadRates fetch (all rates for date) and save if not
func (p Provider) PreloadRates(dateObject time.Time, save bool) (map[string]float64, map[string]float64, time.Time, error) {
	directRates, reverseRates, providerGeneratedTime, err := p.fetchHistoricalRatesAllSymbols(dateObject.Format(util.DateFormatEu))
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	// Save fetched rates to database
	if save {
		p.saveHistoricalRatesAllSymbols("AED", directRates, reverseRates, dateObject, providerGeneratedTime)
	}
	return directRates, reverseRates, providerGeneratedTime, nil
}

// IsRequestValid validates API call to provider.
func (p Provider) IsRequestValid(ratesRequest model.RatesRequest) (bool, error) {
	// BaseProvider API call request validation
	if _, err := p.BaseProvider.IsRequestValid(p, ratesRequest); err != nil {
		return false, err
	}

	// Provider request validation. Check AED is in baseCurrency OR ONLY AED in symbols
	if !(ratesRequest.BaseCurrency == "AED" || (len(ratesRequest.Symbols) == 1 && ratesRequest.Symbols[0] == "AED")) {
		return false, errors.New("provider needs AED is in baseCurrency OR ONLY AED in symbols")
	}
	return true, nil
}

// GetRateGenerationTime returns historical rates generated time on provider side
func (p Provider) GetRateGenerationTime() time.Time {
	return p.BaseProvider.GetRateGenerationTime(p.config.RatesGeneratedTime)
}

// BuildEntity builds entity with given rates
func (p Provider) BuildEntity(endpoint string, baseCurrency string, quotedCurrency string, rate float64, rateDate time.Time, providerDate time.Time) *entity.CurrencyRate {
	e := p.BaseProvider.BuildEntity(endpoint, p.GetCode(), baseCurrency, quotedCurrency, rate, rateDate, providerDate)
	return e
}

// GetLocation returns location for current provider
func (p Provider) GetLocation() *time.Location {
	location, _ := time.LoadLocation("Asia/Dubai")
	return location
}

// GetSupportedCurrencies returns list of currencies, supported by provider
func (p Provider) GetSupportedCurrencies() []string {
	return p.config.SupportedCurrencies
}

// filterRates filter all fetched rates by symbols passed in arguments
func (p Provider) filterRates(rates map[string]float64, baseCurrency string, symbols []string) map[string]float64 {
	var filteredRates = make(map[string]float64)
	for _, symbol := range symbols {
		if symbol == baseCurrency {
			filteredRates[symbol] = 1
		} else {
			filteredRates[symbol] = rates[symbol]
		}
	}
	return filteredRates
}

// fetchHistoricalRatesAllSymbols - fetches directRates and reverseRates. Return normalized (scale=6) rates
func (p Provider) fetchHistoricalRatesAllSymbols(date string) (map[string]float64, map[string]float64, time.Time, error) {
	var (
		directRates  = make(map[string]float64)
		reverseRates = make(map[string]float64)
		providerDate time.Time
		err          error
		resp         *http.Response
		body         []byte
	)

	url := "https://www." + "centralbank" + ".ae" + "/en/fx-rates-ajax?date=" + date + "&v=2"
	if resp, err = http.Get(url); err != nil {
		return nil, nil, time.Time{}, err
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, nil, time.Time{}, err
	}
	if directRates, reverseRates, providerDate, err = p.getRatesFromResponse(body); err != nil {
		return nil, nil, time.Time{}, err
	}
	return directRates, reverseRates, providerDate, nil
}

// saveHistoricalRatesAllSymbols save all history currency rates for one day
func (p Provider) saveHistoricalRatesAllSymbols(
	baseCurrency string,
	directRates map[string]float64,
	reverseRates map[string]float64,
	date time.Time,
	providerDate time.Time) error {
	var e *entity.CurrencyRate

	// Save direct rates
	for quotedCurrency, directRate := range directRates {
		e = p.BuildEntity(util.EndpointHistorical, baseCurrency, quotedCurrency, directRate, date, providerDate)
		p.saveEntity(e)
	}

	// Save reverse rates
	for quotedCurrency, reverseRate := range reverseRates {
		e = p.BuildEntity(util.EndpointHistorical, quotedCurrency, baseCurrency, reverseRate, date, providerDate)
		p.saveEntity(e)
	}
	return nil
}

// saveEntity saves single pair currency rate entity to database
func (p Provider) saveEntity(entity *entity.CurrencyRate) *gorm.DB {
	// OnConflict is needed if there is forward Latest to History(yesterday) endpoint and rates already loaded
	return p.db.Clauses(clause.OnConflict{DoNothing: true}).Create(entity)
}

// getRatesFromResponse parse response and get fetch rates from it
func (p Provider) getRatesFromResponse(body []byte) (map[string]float64, map[string]float64, time.Time, error) {
	var (
		err                    error
		apiJson                ApiResponse
		table                  string
		directRates            = make(map[string]float64)
		reverseRates           = make(map[string]float64)
		normalizedDirectRates  = make(map[string]float64)
		normalizedReverseRates = make(map[string]float64)
		providerGeneratedTime  time.Time
		doc                    *goquery.Document
		currencyRate           float64
		currencyCode           string
	)
	if err = json.Unmarshal(body, &apiJson); err != nil {
		return directRates, reverseRates, time.Time{}, err
	}
	table = apiJson.Table

	// Load the HTML document
	reader := strings.NewReader(table)
	if doc, err = goquery.NewDocumentFromReader(reader); err != nil {
		log.Fatal(err)
		return directRates, reverseRates, time.Time{}, err
	}

	// Find rates
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		tds := s.Find("td")
		currencyName := tds.Eq(0).Text()
		currencyRateStr := tds.Eq(1).Text()
		if currencyRate, err = strconv.ParseFloat(currencyRateStr, 64); err != nil {
			return
		}
		if currencyCode, err = p.getCurrencyCodeByName(currencyName); err != nil {
			return
		}
		reverseRates[currencyCode] = currencyRate
	})

	// Find provider generated date
	dateStr := apiJson.LastUpdated

	// Try to parse date from provider
	for _, format := range p.getDateFormats() {
		if providerGeneratedTime, err = time.ParseInLocation(format, dateStr, p.GetLocation()); err == nil {
			break
		}
	}

	// Centralbank.ae returns reverse rates, need convert to direct
	for cur, reverseRate := range reverseRates {
		normalizedReverseRate := math.Round(reverseRate*1000000) / 1000000
		normalizedReverseRates[cur] = normalizedReverseRate
		normalizedDirectRates[cur] = math.Round((1/normalizedReverseRate)*1000000) / 1000000
	}

	// Can not parse date from rates provider
	return normalizedDirectRates, normalizedReverseRates, providerGeneratedTime, err
}

// getDateFormats returns provider-related date formats
func (p Provider) getDateFormats() []string {
	return []string{
		"02 Jan 2006 3:04 PM",
		"02 Jan 2006 3:04PM",
		"02 Jan 2006 15:04:05 PM",
		"02 Jan 2006 15:04:05PM",
	}
}

// getCurrencyCodeByName return currency code by currency provider related currency name
func (p *Provider) getCurrencyCodeByName(name string) (string, error) {
	nameToCode := map[string]string{
		"US Dollar":               "USD",
		"Argentine Peso":          "ARS",
		"Australian Dollar":       "AUD",
		"Bangladesh Taka":         "BDT",
		"Bahrani Dinar":           "BHD",
		"Brunei Dollar":           "BND",
		"Brazilian Real":          "BRL",
		"Botswana Pula":           "BWP",
		"Belarus Rouble":          "BYN",
		"Canadian Dollar":         "CAD",
		"Swiss Franc":             "CHF",
		"Chilean Peso":            "CLP",
		"Chinese Yuan - Offshore": "CNH",
		"Chinese Yuan":            "CNY",
		"Colombian Peso":          "COP",
		"Czech Koruna":            "CZK",
		"Danish Krone":            "DKK",
		"Algerian Dinar":          "DZD",
		"Egypt Pound":             "EGP",
		"Euro":                    "EUR",
		"GB Pound":                "GBP",
		"Hongkong Dollar":         "HKD",
		"Hungarian Forint":        "HUF",
		"Indonesia Rupiah":        "IDR",
		"Indian Rupee":            "INR",
		"Iceland Krona":           "ISK",
		"Jordan Dinar":            "JOD",
		"Japanese Yen":            "JPY",
		"Kenya Shilling":          "KES",
		"Korean Won":              "KPW",
		"Kuwaiti Dinar":           "KWD",
		"Kazakhstan Tenge":        "KZT",
		"Lebanon Pound":           "LBP",
		"Sri Lanka Rupee":         "LKR",
		"Moroccan Dirham":         "MAD",
		"Macedonia Denar":         "MKD",
		"Mexican Peso":            "MXN",
		"Malaysia Ringgit":        "MYR",
		"Nigerian Naira":          "NGN",
		"Norwegian Krone":         "NOK",
		"NewZealand Dollar":       "NZD",
		"Omani Rial":              "OMR",
		"Peru Sol":                "PEN",
		"Philippine Piso":         "PHP",
		"Pakistan Rupee":          "PKR",
		"Polish Zloty":            "PLN",
		"Qatari Riyal":            "QAR",
		"Serbian Dinar":           "RSD",
		"Russia Rouble":           "RUB",
		"Saudi Riyal":             "SAR",
		"Sudanese Pound":          "SDG",
		"Swedish Krona":           "SEK",
		"Singapore Dollar":        "SGD",
		"Thai Baht":               "THB",
		"Tunisian Dinar":          "TND",
		"Turkish Lira":            "TRY",
		"Trin Tob Dollar":         "TTD",
		"Taiwan Dollar":           "TWD",
		"Tanzania Shilling":       "TZS",
		"Uganda Shilling":         "UGX",
		"Vietnam Dong":            "VND",
		"Yemen Rial":              "YER",
		"South Africa Rand":       "ZAR",
		"Zambian Kwacha":          "ZMW",
		"Azerbaijan manat":        "AZN",
		"Bulgarian lev":           "BGN",
		"Croatian kuna":           "HRK",
		"Ethiopian birr":          "ETB",
		"Iraqi dinar":             "IQD",
		"Israeli new shekel":      "ILS",
		"Libyan dinar":            "LYD",
		"Mauritian rupee":         "MUR",
		"Romanian leu":            "RON",
		"Syrian pound":            "SYP",
		"Turkmen manat":           "TMT",
		"Uzbekistani som":         "UZS",
	}
	code, ok := nameToCode[name]
	if ok {
		return code, nil
	} else {
		return "", errors.New("currency code not found by name" + name)
	}
}
