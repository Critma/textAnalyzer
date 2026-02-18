package analyze

import (
	"github.com/Critma/textAnalyzer/analyzer/internal/cache"
	"github.com/Critma/textAnalyzer/analyzer/internal/config"
	"github.com/Critma/textAnalyzer/analyzer/internal/models"
	"github.com/Critma/textAnalyzer/analyzer/internal/routes"
	"github.com/rs/zerolog/log"
)

// Worker analyze task, send result back
func Worker(jobs <-chan *models.JsonInput, app config.Application) {
	for task := range jobs {
		// Analyze text
		analyze := analyzeText(task.Text)

		// Cache result
		cache.SetInRedis(analyze, app.Redis, task.Text)

		// Send result back
		output := models.JsonRequestOutput{
			ID:      task.ID,
			Status:  string(models.Success),
			Analyze: *analyze,
		}
		if err := routes.SendResult(output, app.HttpClient, app.Config.ReceiverAddr); err != nil {
			log.Error().Str("event", "send result back").Any("obj", output).Err(err).Msg("failed to send result back")
		}
	}
}
