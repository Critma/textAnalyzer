package routes

import (
	"receiver/internal/config"

	"github.com/gin-gonic/gin"
)

type Routes struct {
	App config.Application
}

func New(app config.Application) Routes {
	return Routes{
		App: app,
	}
}

// Mount connect routes to group
func (r *Routes) Mount(rg *gin.RouterGroup) {
	router := rg.Group("/v1")
	{
		router.POST("/text", r.handleCreate)
		router.GET("/status/:id", r.getStatus)
		router.GET("/health", r.healthCheck)

		router.POST("/result", r.updateAnalyze)
	}
}
