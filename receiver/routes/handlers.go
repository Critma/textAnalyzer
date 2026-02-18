package routes

import (
	"context"
	"net/http"
	"receiver/cache"
	"receiver/internal/metrics"
	"receiver/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// @Summary Create new text analysis request
// @Description Creates a new text analysis request by accepting user-submitted text for processing.
// @Tags Requests
// @Accept json
// @Produce json
// @Param payload body JsonTextInput true "Input text to create an analysis request"
// @Success 200 {object} IdResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /text [post]
func (r *Routes) handleCreate(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.ObserveRequest(time.Since(start), c.Writer.Status(), "handleCreate")
	}()
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

// @Summary Get status of a specific text analysis request
// @Description Retrieves the current status of a text analysis request based on its unique identifier.
// @Tags Requests
// @Accept json
// @Produce json
// @Param id path string true "Unique Request ID"
// @Success 200 {object} JsonRequest
// @Success 200 {object} JsonStatusOnlyOutput
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /status/{id} [get]
func (r *Routes) getStatus(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.ObserveRequest(time.Since(start), c.Writer.Status(), "getStatus")
	}()
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

// @Summary Ping Redis connection for health check
// @Description Performs a simple ping operation against Redis to verify connectivity.
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 503 {object} StatusResponse
// @Router /health [get]
func (r *Routes) healthCheck(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.ObserveRequest(time.Since(start), c.Writer.Status(), "health")
	}()
	// Check Redis
	if err := r.App.Redis.Ping(context.Background()).Err(); err != nil {
		log.Error().Str("handler", "health check").Err(err).Msg("redis ping error")
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "component": "redis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// @Summary Update analyzed text request
// @Description Updates the status and analysis results of a given text request using information received from the analyzer service.
// @Tags Microservices
// @Accept json
// @Produce json
// @Param payload body JsonRequest true "Updated request details including analysis results"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /result [put]
func (r *Routes) updateAnalyze(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.ObserveRequest(time.Since(start), c.Writer.Status(), "updateAnalyze")
	}()
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
