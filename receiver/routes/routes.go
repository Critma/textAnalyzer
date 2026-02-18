package routes

import (
	"receiver/internal/config"
	"time"

	_ "receiver/cmd/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Routes struct {
	App config.Application
}

func New(app config.Application) Routes {
	return Routes{
		App: app,
	}
}

// Mount connect routes(handlers) and middleware to group
func (r *Routes) Mount(rg *gin.Engine) {
	// cross config
	rg.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://" + r.App.Config.Addr, "http://127.0.0.1:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// prometheus metrics
	rg.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router := rg.Group("/api")
	router = router.Group("/v1")
	{
		router.POST("/text", r.handleCreate)
		router.GET("/status/:id", r.getStatus)
		router.GET("/health", r.healthCheck)

		router.POST("/result", r.updateAnalyze)

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
