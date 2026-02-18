package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Critma/textAnalyzer/analyzer/internal/cache"
	"github.com/Critma/textAnalyzer/analyzer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// healthCheck ping redis, return status ok or unavailable
func (r *Routes) healthCheck(c *gin.Context) {
	// Check Redis
	if err := r.App.Redis.Ping(context.Background()).Err(); err != nil {
		log.Error().Str("handler", "health check").Err(err).Msg("redis ping error")
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "component": "redis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *Routes) handleAnalyze(c *gin.Context) {
	var input models.JsonInput
	// parse request body
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		log.Error().Str("handler", "handle analyze").Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if input.Text == "" {
		log.Error().Str("handler", "handle analyze").Msg("Text cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text cannot be empty"})
		return
	}

	// check if request cached, if not continue
	cachedResult := cache.GetFromRedis(r.App.Redis, input.Text)
	if cachedResult != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Success", "cached": true})
		log.Info().Str("text", input.Text).Msg("use cache")
		SendResult(models.JsonRequestOutput{
			ID:      input.ID,
			Status:  string(models.Success),
			Analyze: *cachedResult,
		}, r.App.HttpClient, r.App.Config.ReceiverAddr)
		return
	}

	// try to send task to analyze workers within timeout or ignore
	select {
	case r.Jobs <- &input:
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	case <-time.After(5 * time.Second):
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "Unable to send job within timeout"})
	}
}

// SendResult send JsonRequestOutput to receiver service
func SendResult(output models.JsonRequestOutput, httpClient *http.Client, receiverAddr string) error {
	data, _ := json.Marshal(output)
	addr := fmt.Sprintf("http://%s/api/v1/result", receiverAddr)
	resp, err := httpClient.Post(addr, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("receiver service returned status %d", resp.StatusCode)
	}

	return nil
}
