package routes

import (
	"github.com/Critma/textAnalyzer/analyzer/internal/config"
	"github.com/Critma/textAnalyzer/analyzer/internal/models"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	App  config.Application
	Jobs chan *models.JsonInput
}

func New(app config.Application, jobs chan *models.JsonInput) Routes {
	return Routes{
		App:  app,
		Jobs: jobs,
	}
}

// Mount connect routes to routerGroup
func (r *Routes) Mount(rg *gin.RouterGroup) {
	router := rg.Group("/v1")
	{
		router.POST("/analyze", r.handleAnalyze)
		router.GET("/health", r.healthCheck)
	}
}
