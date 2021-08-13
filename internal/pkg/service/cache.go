package service

import (
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
	cache_store "github.com/netandreus/go-forex-rates/internal/pkg/cache/store"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	gocache "github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"time"
)

// BuildCache /* *cache.ChainCache
func BuildCache(config *model.ApplicationConfig, mysqlClient *gorm.DB) (*cache.ChainCache, error) {
	gocacheClient := gocache.New(1*time.Second, 1*time.Second) // 600 sec for production
	gocacheStore := store.NewGoCache(gocacheClient, nil)
	mysqlStore := cache_store.NewMySQLStore(mysqlClient, nil)
	// Initialize chained cache
	cacheManager := cache.NewChain(
		cache.New(gocacheStore),
		cache.New(mysqlStore),
	)
	return cacheManager, nil
}
