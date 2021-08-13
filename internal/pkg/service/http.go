package service

import (
	"github.com/gin-gonic/gin"
	"github.com/netandreus/go-forex-rates/internal/pkg/controller"
	"github.com/netandreus/go-forex-rates/internal/pkg/model"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

// BuildHttp /* *gin.Engine
func BuildHttp(apiController *controller.ApiController, config *model.ApplicationConfig) (*gin.Engine, error) {
	// Settings
	gin.SetMode(config.Engine.Mode)

	r := gin.Default()
	// Home page
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Currency rates microservice. Read the documentation to learn the API.",
		})
	})
	// Swagger
	url := ginSwagger.URL("/api/v1/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.GET("/api/doc", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	// API
	v1 := r.Group("/api/v1")
	{
		// API doc: api.doc and api.yaml
		v1.StaticFile("/doc.json", "./api/swagger.json")
		v1.StaticFile("/doc.yaml", "./api/swagger.yaml")

		// Health check
		v1.GET("/status", apiController.Status())

		// Historical endpoint
		v1.GET("/historical/:provider/:date", apiController.Historical())

		// Latest endpoint
		v1.GET("/latest/:provider", apiController.Latest())
	}

	return r, nil
}
