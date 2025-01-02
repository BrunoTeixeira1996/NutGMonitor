package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	// telegram bot IP
	resp, err := http.Post("http://192.168.30.21:8000/forward", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	return nil
}
