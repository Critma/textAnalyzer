package analyze

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/Critma/textAnalyzer/analyzer/internal/models"
)

func analyzeText(text string) *models.JsonAnalyze {
	// Count characters
	charCount := len(text)

	// Count words
	words := strings.Fields(text)
	wordCount := len(words)

	// Count sentences
	sentenceRegex := regexp.MustCompile(`[.!?]+`)
	sentences := sentenceRegex.Split(text, -1)
	// Filter out empty strings
	sentenceCount := 0
	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			sentenceCount++
		}
	}

	// Calculate average word length
	var totalWordLength int
	for _, word := range words {
		// Remove punctuation from word
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if word != "" {
			totalWordLength += len(word)
		}
	}
	var averageWordLength float64
	if wordCount > 0 {
		averageWordLength = float64(totalWordLength) / float64(wordCount)
	}

	return &models.JsonAnalyze{
		WordCount:         wordCount,
		CharCount:         charCount,
		SentenceCount:     sentenceCount,
		AverageWordLength: averageWordLength,
	}
}
