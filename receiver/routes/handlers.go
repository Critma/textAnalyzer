package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"receiver/cache"
	"receiver/internal/config"
	"receiver/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (r *Routes) handleCreate(c *gin.Context) {
	var req JsonTextInput
	//TODO limit read body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Str("handler", "handle request").Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Text == "" {
		log.Error().Str("handler", "handle request").Msg("Text cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text cannot be empty"})
		return
	}

	// Save to storage
	request, err := r.App.Store.Requests.CreateRequest(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save text"})
		return
	}

	// Send to analyzer service
	if err := sendToAnalyzer(r.App, request.ID, req.Text); err != nil {
		// Update status to failed
		r.App.Store.Requests.UpdateRequest(request.ID, storage.Failed, storage.AnalyzeResult{})
		log.Error().Str("handler", "handle request").Err(err).Msg("Failed to send to analyzer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send to analyzer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": request.ID})
}

func sendToAnalyzer(app config.Application, id uuid.UUID, text string) error {
	var toSend *JsonToAnalyzer = &JsonToAnalyzer{ID: id.String(), Text: text}
	data, err := json.Marshal(toSend)
	if err != nil {
		return err
	}

	analyzerUrl := fmt.Sprintf("http://%s/api/v1/analyze", app.Config.AnalyzerAddr)
	req, err := http.NewRequest("POST", analyzerUrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// send to analyzer service
	resp, err := app.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("analyzer service returned status %d", resp.StatusCode)
	}

	return nil
}

func (r *Routes) getStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Error().Str("handler", "get status").Err(err).Msg("Invalid ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var result *storage.TextRequest
	// check is request cached
	result = cache.GetFromRedis(r.App.Redis, id)

	if result == nil {
		// request not cached, get from store
		result, err = r.App.Store.Requests.GetRequest(id)
		if err != nil {
			if err == storage.ErrNotFound {
				log.Error().Str("handler", "get status").Str("requestID", id.String()).Err(err).Msg("request not found")
				c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
				return
			}
			log.Error().Str("handler", "get status").Str("requestID", id.String()).Err(err).Msg("found request error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
			return
		}
		cache.SetInRedis(result, r.App.Redis, id)
	}

	// success: full result, else only status
	switch result.Status {
	case storage.InProcess, storage.Failed:
		c.JSON(http.StatusOK, JsonStatusOnlyOutput{ID: result.ID, Text: result.Text, Status: result.Status})
	case storage.Success:
		c.JSON(http.StatusOK, JsonRequest{
			ID:     result.ID,
			Text:   result.Text,
			Status: result.Status,
			Analyze: JsonAnalyze{
				WordCount:         result.Analyze.WordCount,
				CharCount:         result.Analyze.CharCount,
				SentenceCount:     result.Analyze.SentenceCount,
				AverageWordLength: result.Analyze.AverageWordLength,
			},
		})
	}
}

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

func (r *Routes) updateAnalyze(c *gin.Context) {
	//TODO receive only service ip
	var answer JsonRequest
	if err := c.ShouldBindJSON(&answer); err != nil {
		log.Error().Str("handler", "update analyze").Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	analyzeResult := storage.AnalyzeResult{
		WordCount:         answer.Analyze.WordCount,
		CharCount:         answer.Analyze.CharCount,
		SentenceCount:     answer.Analyze.SentenceCount,
		AverageWordLength: answer.Analyze.AverageWordLength,
	}
	result, err := r.App.Store.Requests.UpdateRequest(answer.ID, answer.Status, analyzeResult)
	if err != nil {
		if err == storage.ErrNotFound {
			log.Error().Str("handler", "update analyze").Str("requestID", answer.ID.String()).Err(err).Msg("request not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}
		log.Error().Str("handler", "update analyze").Str("requestID", answer.ID.String()).Err(err).Msg("request update error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request update error"})
		return
	}
	cache.SetInRedis(result, r.App.Redis, result.ID)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
