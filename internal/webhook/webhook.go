package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/alertmanager"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/forward"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/targets"
)

type Hooker struct {
	Targets     []targets.Target
	mu          sync.Mutex
	firingCount int
	cancel      context.CancelFunc
}

// sendTelegramAsync sends a message to Telegram asynchronously with timeout
// This prevents blocking the webhook if Telegram is unreachable
func (h *Hooker) sendTelegramAsync(status string, messageContent string, structToSend interface{}, messageErr string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- forward.ForwardMessageToTelegram(status, messageContent, structToSend, messageErr)
		}()

		select {
		case <-ctx.Done():
			logger.Log.Printf("[webhook warning] telegram timeout for status: %s\n", status)
		case err := <-errChan:
			if err != nil {
				logger.Log.Printf("[webhook warning] telegram error: %s\n", err)
			}
		}
	}()
}

// FIXME: move the action of checking status and powering off to another folder outside webhook
func (h *Hooker) alertmanager(w http.ResponseWriter, r *http.Request) {
	var (
		alertmanagerResponse alertmanager.AlertManagerResponse
	)

	defer r.Body.Close()
	if r.Method != "POST" {
		http.Error(w, "NOT POST!", http.StatusBadRequest)
		logger.Log.Printf("[webhook error] received a request that was not POST instead it was %s\n", r.Method)
		return
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&alertmanagerResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Printf("[webhook error] while unmarshal json response %s\n", err)
		return
	}

	logger.Log.Printf("RECEIVED ALERT FROM ALERT MANAGER: %v\n", alertmanagerResponse)

	// Respond immediately to Alert Manager to avoid timeout
	w.WriteHeader(http.StatusOK)

	// Lock for state management
	h.mu.Lock()
	defer h.mu.Unlock()

	// when there's a new alert with nut_status == 0 (Unknown)
	if alertmanagerResponse.Alerts[0].Labels.Alertname == "UPSStatusUnknown" {
		res := "unknown state on ups"
		logger.Log.Printf("[webhook info] %s\n", res)
		h.sendTelegramAsync("NOT OK", res, alertmanagerResponse, "maybe nut ups docker container broke")
		h.cancel() // terminate the webhook
	}

	logger.Log.Printf("alertmanagerResponseStatus: %s, alertmanagerResponseAlertsLabelsAlertname: %s\n", alertmanagerResponse.Status, alertmanagerResponse.Alerts[0].Labels.Alertname)

	// when there's a new alert with nut_status == 2 (OB)
	if alertmanagerResponse.Status == "firing" && alertmanagerResponse.Alerts[0].Labels.Alertname == "UPSStatusCritical" {
		h.firingCount++

		// After 3 attempts we start to shutdown (3 attempts +/- 5 minutes)
		if h.firingCount >= 3 {
			res := "firing count happened 3 times (waited around 5 minutes) ... shutting down targets and telegram bot will stop here"
			logger.Log.Printf("[webhook info] %s\n", res)
			h.sendTelegramAsync("SHUTDOWN ACTION", res, alertmanagerResponse, "")
			h.cancel() // terminate the webhook

		} else {
			res := fmt.Sprintf("firing count: %d", h.firingCount)
			logger.Log.Printf("[webhook info] %s\n", res)
			h.sendTelegramAsync("FIRING ACTION", res, alertmanagerResponse, "")
		}
	} else {
		res := fmt.Sprintf("alert was resolved with firing count: %d", h.firingCount)
		logger.Log.Printf("[webhook info] %s\n", res)
		h.sendTelegramAsync("RESOLVED ACTION", res, alertmanagerResponse, "")

		h.firingCount = 0
	}
}

func StartWebHook(upsTargets []targets.Target) error {
	ctx, cancel := context.WithCancel(context.Background())

	hooker := &Hooker{Targets: upsTargets, cancel: cancel}
	http.HandleFunc("/alertmanager", hooker.alertmanager)

	logger.Log.Println("[webhook info] listening on port 9999")

	server := &http.Server{Addr: ":9999"}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("[webhook error] error listening: %s\n", err)
		}
	}()

	// Wait until webhook is canceled by alert logic (firing count >= 3)
	// The AlertFastPowerOff() goroutine will also trigger shutdown if UPS is on battery for 3+ minutes
	<-ctx.Done()
	logger.Log.Println("[webhook info] webhook canceled by alert logic")

	// we want to shutdown all targets
	targets.ShutdownTargets(hooker.Targets)

	logger.Log.Printf("[webhook info] checking target status\n")
	downTargets := targets.CheckTargetsStatus(hooker.Targets)

	res := fmt.Sprintf("shut down targets: %v size (%d) - all targets size: (%d)\n", downTargets, len(downTargets), len(hooker.Targets)-1)
	logger.Log.Printf("[webhook info] %s\n", res)

	// initiate graceful shutdown
	logger.Log.Println("[webhook info] shutting down server...")

	// wait 1 minute and dont accept any more requests
	// immediate shutdown if there are no jobs pending
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatalf("[webhook error] server shutdown failed: %s", err)
	}
	logger.Log.Println("[webhook info] server shut down gracefully")

	return nil
}