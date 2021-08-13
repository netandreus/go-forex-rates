package service

import (
	"github.com/gocolly/colly/v2"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"time"
)

// BuildColly /* *colly.Collector
func BuildColly(config *model.ApplicationConfig) (*colly.Collector, error) {
	coll := colly.NewCollector(
		colly.MaxDepth(1),
	//colly.Async(true), // does not working..
	)
	coll.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Collector.Parallelism,
		RandomDelay: time.Duration(config.Collector.RandomDelay) * time.Second,
	})
	return coll, nil
}
