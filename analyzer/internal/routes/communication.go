package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Critma/textAnalyzer/analyzer/internal/models"
)

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
