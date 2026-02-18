package models

type Status string

var (
	InProcess Status = "in process"
	Success   Status = "success"
	Failed    Status = "failed"
)

type JsonInput struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type JsonRequestOutput struct {
	ID      string      `json:"id"`
	Status  string      `json:"status"`
	Analyze JsonAnalyze `json:"analyze"`
}

type JsonAnalyze struct {
	WordCount         int     `json:"wordCount,omitempty"`
	CharCount         int     `json:"charCount,omitempty"`
	SentenceCount     int     `json:"sentenceCount,omitempty"`
	AverageWordLength float64 `json:"averageWordLength,omitempty"`
}
