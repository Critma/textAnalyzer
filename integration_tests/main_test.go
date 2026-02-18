// integration_test.go
package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"net/http"
	"testing"
)

const (
	receiverHost = "http://localhost:8080"
	analyzerHost = "http://localhost:8081"
)

func TestHealthChecks(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{"Receiver Health Check", receiverHost + "/api/v1/health"},
		{"Analyzer Health Check", analyzerHost + "/api/v1/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.endpoint, nil)
			if err != nil {
				t.Fatalf("couldn't create request: %v", err)
			}

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to do request: %v", err)
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if res.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK but got %d (%s)", res.StatusCode, string(body))
			}
		})
	}
}

func TestEndToEndFlow(t *testing.T) {
	testCases := []struct {
		name                      string
		text                      string
		expectedWordCount         int
		expectedCharCount         int
		expectedSentenceCount     int
		expectedAverageWordLength float64
	}{
		{"Simple Text Analysis", "Hello, how are you?", 4, 19, 1, 3.5},
		{"Complex Text Analysis", "This is a complex sentence. It has multiple sentences. Each one needs counting.", 13, 79, 3, 4.92},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// send to receiver
			postEndpoint := receiverHost + "/api/v1/text"
			data := bytes.NewBuffer(fmt.Appendf(nil, `{"text": "%s"}`, tc.text))
			req, err := http.NewRequest("POST", postEndpoint, data)
			if err != nil {
				t.Fatalf("can't create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}
			defer res.Body.Close()

			// check success
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Expected 201 created but got %d", res.StatusCode)
			}

			// parse id from response
			var response struct {
				ID string `json:"id"`
			}
			err = json.NewDecoder(res.Body).Decode(&response)
			if err != nil {
				t.Fatalf("cannot decode response: %v", err)
			}

			// wait until processed
			time.Sleep(2 * time.Second)

			// check analyze result
			getEndpoint := fmt.Sprintf(receiverHost+"/api/v1/status/%s", response.ID)
			reqGet, err := http.NewRequest("GET", getEndpoint, nil)
			if err != nil {
				t.Fatalf("can't create GET request: %v", err)
			}

			resGet, err := client.Do(reqGet)
			if err != nil {
				t.Fatalf("failed to execute GET request: %v", err)
			}
			defer resGet.Body.Close()

			if resGet.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 OK but got %d", resGet.StatusCode)
			}

			// parse
			type analysis struct {
				WordsCount     int     `json:"wordCount"`
				CharsCount     int     `json:"charCount"`
				SentencesCount int     `json:"sentenceCount"`
				AverageWordLen float64 `json:"averageWordLength"`
			}
			var wrapper struct {
				Analyze analysis `json:"analyze"`
			}
			err = json.NewDecoder(resGet.Body).Decode(&wrapper)
			if err != nil {
				t.Fatalf("cannot decode analysis response: %v", err)
			}

			// check expexted and actual
			if tc.expectedWordCount != wrapper.Analyze.WordsCount ||
				tc.expectedCharCount != wrapper.Analyze.CharsCount ||
				tc.expectedSentenceCount != wrapper.Analyze.SentencesCount ||
				math.Round(tc.expectedAverageWordLength) != math.Round(wrapper.Analyze.AverageWordLen) {
				t.Fatalf("Analysis results don't match expectations:\nGot %+v\nWant %+v", wrapper.Analyze, struct {
					WordsCount     int
					CharsCount     int
					SentencesCount int
					AverageWordLen float64
				}{tc.expectedWordCount, tc.expectedCharCount, tc.expectedSentenceCount, tc.expectedAverageWordLength})
			}
		})
	}
}
