package service

import (
	"github.com/gin-gonic/gin"
	"github.com/netandreus/go-forex-rates/internal/pkg/logger"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strconv"
)

// BuildDatabase /** *model.ApplicationConfig
func BuildDatabase(config *model.ApplicationConfig) (*gorm.DB, error) {
	db, err := initDatabase(
		config.L2Cache.Hostname,
		config.L2Cache.Port,
		config.L2Cache.Username,
		config.L2Cache.Password,
		config.L2Cache.Database)
	if err != nil {
		logger.LogError(err.Error(), "DB")
		os.Exit(0)
	}
	// Mode (debug / release)
	if config.Engine.Mode == gin.DebugMode {
		db = db.Debug()
	}
	return db, nil
}

func initDatabase(host string, port int, login string, password string, database string) (db *gorm.DB, err error) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := login + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.LogError(err.Error(), "DB")
		os.Exit(0)
	}
	// @todo Use connection pool and close it in closeHandler
	// @see https://gorm.io/docs/connecting_to_the_database.html#Connection-Pool
	return db, err
}
