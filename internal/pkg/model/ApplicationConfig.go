package model

// ApplicationConfig represents structure of main application config
type ApplicationConfig struct {
	// HTTP server (Gin) settings
	Engine EngineConfig `yaml:"engine"`

	// Collector settings
	Collector struct {
		// Number of parallel parser workers
		Parallelism int `yaml:"parallelism" env:"COLLECTOR_PARALLELISM" env-default:4`

		// Delay between requests of provider API
		RandomDelay int `yaml:"random_delay" env:"COLLECTOR_DELAY" env-default:1`
	} `yaml:"collector"`

	// Level-1 cache settings (go-cache)
	L1Cache struct {
		// Cache TTL
		DefaultExpiration int `yaml:"default_expiration" env:"L1_DEFAULT_EXPIRATION" env-default:30`

		// Cache cleanup interval (gc calling timeout)
		CleanupInterval int `yaml:"cleanup_interval" env:"L1_CLEANUP_INTERVAL" env-default:60`
	} `yaml:"l1_cache"`

	// Level-2 cache settings (MySQL)
	L2Cache struct {
		// Database server hostname
		Hostname string `yaml:"hostname" env:"L2_HOSTNAME" env-default:"127.0.0.1"`

		// Database server port
		Port int `yaml:"port" env:"L2_PORT" env-default:3306`

		// Database server username
		Username string `yaml:"username" env:"L2_USERNAME"`

		// Database server password
		Password string `yaml:"password" env:"L2_PASSWORD"`

		// Database server database name
		Database string `yaml:"database" env:"L2_DATABASE" env-default:"go_forex_rates"`
	} `yaml:"l2_cache"`

	// Providers settings
	Providers map[string]ProviderConfig
}

// EngineConfig is HTTP server settings
type EngineConfig struct {
	// Running mode: debug / release
	Mode string `yaml:"mode" env:"mode" env-default:"debug"`

	// Listen port
	Port int `yaml:"port" env:"PORT" env-default:9090`
}

// ProviderConfig is configuration of currency rates provider
type ProviderConfig struct {
	// Time location of provider's rates generating center
	Location string `yaml:"location" env-default:""UTC`

	// Time, when provider generate historical rates for today
	RatesGeneratedTime string `yaml:"rates_generated_time" env-default:"23:59:59"`

	// Access token for provider's API
	APIKey string `yaml:"api_key" env-default:""`

	// List of currencies, supporting by provider
	SupportedCurrencies []string `yaml:"supported_currencies"`

	// Enable or disable preload historical rates to L2 cache (database)
	HistoricalPreload bool `yaml:"historical_preload"`

	// Start date for preload historical currency rates
	HistoricalStartDate string `yaml:"historical_start_date"`
}
