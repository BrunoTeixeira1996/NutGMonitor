package ups

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestValidateNutUPSContainer_ValidResponse only exercises the success path,
// which never sends anything to Telegram. The error paths call
// forward.ForwardMessageToTelegram directly, so they're intentionally not
// covered here to keep this test suite free of any real network side effects.
func TestValidateNutUPSContainer_ValidResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	if err := ValidateNutUPSContainer(ts.URL); err != nil {
		t.Errorf("expected no error for a valid response, got %v", err)
	}
}

// onBatteryLines builds n consecutive [OB] log lines. isUPSOnBattery only
// looks at the 6th field, so the rest of the line's content doesn't matter.
func onBatteryLines(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "2024-10-20 23:46:00 100 239.5 6 [OB]"
	}
	return lines
}

func TestOutageDuration(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{
			name: "normal behaviour",
			lines: []string{
				"2024-10-20 22:00:00 100 239.2 4 [OL]",
				"2024-10-20 22:00:02 100 239.2 4 [OL]",
				"2024-10-20 22:00:04 100 239.2 4 [OL]",
			},
			want: 0, // 0 on battery
		},
		{
			name: "fast poweroff",
			lines: []string{
				"2024-10-20 23:45:55 100 239.2 4 [OL]",
				"2024-10-20 23:45:57 100 239.5 6 [OB]",
				"2024-10-20 23:45:59 100 239.5 6 [OB]",
				"2024-10-20 23:46:01 100 239.5 6 [OB]",
				"2024-10-20 23:46:03 100 239.2 4 [OL]",
			},
			want: 3, // 3 on battery
		},
		{
			name:  "long poweroff - starts turning off everything",
			lines: onBatteryLines(90),
			want:  90,
		},
		{
			// Power restoring resets tracking (see AlertFastPowerOff), so two
			// separate blips must not add up - only the longest single one
			// counts towards the poweroff threshold.
			name: "multiple separate fast poweroffs - only the longest one counts",
			lines: []string{
				"2024-10-20 23:40:00 100 239.2 4 [OB]",
				"2024-10-20 23:40:02 100 239.2 4 [OB]",
				"2024-10-20 23:40:04 100 239.2 4 [OL]",
				"2024-10-20 23:45:00 100 239.2 4 [OB]",
				"2024-10-20 23:45:02 100 239.2 4 [OB]",
				"2024-10-20 23:45:04 100 239.2 4 [OB]",
				"2024-10-20 23:45:06 100 239.2 4 [OB]",
				"2024-10-20 23:45:08 100 239.2 4 [OB]",
				"2024-10-20 23:45:10 100 239.2 4 [OL]",
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := outageDuration(tt.lines); got != tt.want {
				t.Errorf("outageDuration() = %d, want %d", got, tt.want)
			}
		})
	}
}

// outageDuration returns how many [OB] log lines in a row the UPS stayed on
// battery for - the same signal AlertFastPowerOff uses (90 lines, at
// upslog's 2s interval, is the ~3 minute sustained-outage threshold).
func outageDuration(lines []string) int {
	longest, current := 0, 0
	for _, l := range lines {
		if isUPSOnBattery(l) {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 0
		}
	}
	return longest
}
