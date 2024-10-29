package targets

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

// Since gokrazy does not support shutting down the system via SSH, we use an HTTP-based approach by making a POST request to the endpoint http://brun0-pi/poweroff. This endpoint triggers the shutdown of the system. Once all services running on gokrazy have been successfully terminated, the server responds with a confirmation, including the duration it took to complete the poweroff process.

func ShutdownGokrazy(sshkey, ip string) error {
	url := "http://" + ip + "/poweroff"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("[gokrazy error] could not create request: %s\n", err)
	}

	gokrazyToken := os.Getenv("GKTOKEN")
	if gokrazyToken == "" {
		return fmt.Errorf("[gokrazy error] GKTOKEN not set so the shutdown was not performed\n")
	}

	req.SetBasicAuth("gokrazy", gokrazyToken)

	// gokrazy should exit in less than 4 minutes, normaly it takes around 1/2 min
	client := &http.Client{Timeout: 4 * time.Minute}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[gokrazy error] could not make request: %s\n", err)
	}
	defer resp.Body.Close()

	logger.Log.Printf("[gokrazy info] response status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[gokrazy error] could not read response body:: %s\n", err)
	}
	logger.Log.Printf("[gokrazy info] response body: %s\n", string(body))

	return nil
}
