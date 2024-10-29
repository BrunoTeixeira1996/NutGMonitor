package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func ForwardMessage(status string, res string, structToSend interface{}) error {
	var message string

	if structToSend == nil {
		message = fmt.Sprintf("%s - %s", status, res)
	} else {
		alertJSON, err := json.Marshal(structToSend)
		if err != nil {
			return fmt.Errorf("[forward error] error marshalling structToSend to JSON: %s\n", err)
		}

		message = fmt.Sprintf("%s - %s - %s", status, res, string(alertJSON))
	}

	requestBody := map[string]string{
		"type":    "nut-alert",
		"message": message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("[forward error] could not marshall JSON: %s\n", err)
	}

	// telegram bot IP
	resp, err := http.Post("http://192.168.30.21:8000/fwd", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	return nil
}
