package forward

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ForwardMessageToTelegram(status string, messageContent string, structToSend interface{}, messageErr string) error {
	var m string

	if structToSend == nil {
		m = fmt.Sprintf("%s - %s", status, messageContent)
	} else {
		alertJSON, err := json.Marshal(structToSend)
		if err != nil {
			return fmt.Errorf("[forward error] error marshalling structToSend to JSON: %s\n", err)
		}

		m = fmt.Sprintf("%s - %s - %s", status, messageContent, string(alertJSON))
	}

	requestBody := map[string]string{
		"source":  "nut-alert",
		"message": m,
		"error":   messageErr,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("[forward error] could not marshall JSON: %s\n", err)
	}

	// Create context with timeout for the HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", "http://bot.lan:8000/forward", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not create request: %s\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// telegram bot IP - use the context-aware client
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request to bot.lan:8000: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[forward error] telegram bot returned status: %d\n", resp.StatusCode)
	}

	return nil
}
