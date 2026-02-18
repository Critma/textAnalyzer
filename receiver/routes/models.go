package routes

import (
	"receiver/internal/storage"

	"github.com/google/uuid"
)

type JsonRequest struct {
	ID      uuid.UUID      `json:"id"`
	Text    string         `json:"text,omitempty"`
	Status  storage.Status `json:"status"`
	Analyze JsonAnalyze    `json:"analyze,omitempty"`
}

type JsonAnalyze struct {
	WordCount         int     `json:"wordCount,omitempty"`
	CharCount         int     `json:"charCount,omitempty"`
	SentenceCount     int     `json:"sentenceCount,omitempty"`
	AverageWordLength float64 `json:"averageWordLength,omitempty"`
}

type JsonStatusOnlyOutput struct {
	ID     uuid.UUID      `json:"id"`
	Text   string         `json:"text"`
	Status storage.Status `json:"status"`
}

type JsonTextInput struct {
	Text string `json:"text"`
}

type JsonToAnalyzer struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
