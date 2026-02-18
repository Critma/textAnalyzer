package storage

import "github.com/google/uuid"

type TextRequest struct {
	ID      uuid.UUID
	Text    string
	Status  Status
	Analyze AnalyzeResult
}

type AnalyzeResult struct {
	WordCount         int
	CharCount         int
	SentenceCount     int
	AverageWordLength float64
}

type Status string

var (
	InProcess Status = "in process"
	Success   Status = "success"
	Failed    Status = "failed"
)
