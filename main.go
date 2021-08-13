// @title Go-forex-rates HTTP REST API server for currency exchange rates
// @version 1.0
// @description Microservice for obtaining exchange rates
// @contact.name API Support
// @contact.email netandreus@gmail.com
// @license.name MIT
// @license.url https://github.com/netandreus/go-forex-rates/blob/master/LICENSE
// @BasePath /api/v1
package main

import (
	"github.com/netandreus/go-forex-rates/api" // swagger docs.go
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider/emirates"
	"github.com/netandreus/go-forex-rates/internal/pkg/provider/fixer"
	"github.com/netandreus/go-forex-rates/pkg/server"
	"gorm.io/gorm"
	"strconv"
)

// global application container
var srv *server.Server

// init initialize server
func init() {
	var err error

	// Build srv (container, services etc)
	srv, err = server.New()
	if err != nil {
		panic("error loading srv: " + err.Error())
		return
	}

	// Add rates providers
	srv.ContainerInvoke(func(registry *provider.Registry, db *gorm.DB, config *model.ApplicationConfig) {
		registry.AddProvider(emirates.New(db, config))
		registry.AddProvider(fixer.New(db, config))
	})
}

// main start srv
func main() {
	var err error
	api.SwaggerInfo.BasePath = "/api/doc"
	api.SwaggerInfo.Host = ":" + strconv.Itoa(srv.GetListenPort())
	if err = srv.Run(); err != nil {
		panic("error starting srv: " + err.Error())
		return
	}
	select {}
}
