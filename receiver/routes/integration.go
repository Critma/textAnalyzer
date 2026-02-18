package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"receiver/internal/config"

	"github.com/google/uuid"
)

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
