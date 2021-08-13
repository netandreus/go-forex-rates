package service

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/netandreus/go-forex-rates/internal/pkg/logger"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"os"
)

// BuildConfig /* *model.ApplicationConfig
func BuildConfig() (*model.ApplicationConfig, error) {
	var (
		config model.ApplicationConfig
		err    error
	)
	err = cleanenv.ReadConfig("./configs/config.yml", &config)
	if err != nil {
		logger.LogError(err.Error(), "CONFIG")
		os.Exit(0)
	}
	return &config, nil
}
