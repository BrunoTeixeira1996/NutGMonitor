package alertmanager

import "time"

type AlertManagerResponse struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Labels struct {
			Alertname string `json:"alertname"`
			Instance  string `json:"instance"`
			Job       string `json:"job"`
			Severity  string `json:"severity"`
			Ups       string `json:"ups"`
		} `json:"labels"`
		StartsAt     time.Time `json:"startsAt"`
		EndsAt       time.Time `json:"endsAt"`
		GeneratorURL string    `json:"generatorURL"`
	} `json:"alerts"`
}

// ON UPS FAILURE
// {"receiver":"webhook","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"UPSStatusCritical","instance":"","job":"nut","severity":"critical","ups":"ups"},"annotations":{"description":"The UPS status is currently critical (status = 2).","summary":"UPS status is critical and its running on battery"},"startsAt":"2024-09-23T16:40:47.492Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":""}],"groupLabels":{},"commonLabels":{"alertname":"UPSStatusCritical","instance":"","job":"nut","severity":"critical","ups":"ups"},"commonAnnotations":{"description":"The UPS status is currently critical (status = 2).","summary":"UPS status is critical and its running on battery"},"externalURL":"","version":"4","groupKey":"{}:{}","truncatedAlerts":0}

// ON UPS RESOLVED
// {"receiver":"webhook","status":"resolved","alerts":[{"status":"resolved","labels":{"alertname":"UPSStatusCritical","instance":"","job":"nut","severity":"critical","ups":"ups"},"annotations":{"description":"The UPS status is currently critical (status = 2).","summary":"UPS status is critical and its running on battery"},"startsAt":"2024-09-23T16:53:32.492Z","endsAt":"2024-09-23T17:00:02.492Z","generatorURL":"","fingerprint":""}],"groupLabels":{},"commonLabels":{"alertname":"UPSStatusCritical","instance":"","job":"nut","severity":"critical","ups":"ups"},"commonAnnotations":{"description":"The UPS status is currently critical (status = 2).","summary":"UPS status is critical and its running on battery"},"externalURL":"","version":"4","groupKey":"{}:{}","truncatedAlerts":0}
