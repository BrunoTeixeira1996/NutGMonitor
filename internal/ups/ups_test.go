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

func TestLongestOnBatteryStreak(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{
			name: "normal - no outage",
			lines: []string{
				"2024-10-20 22:00:00 100 239.2 4 [OL]",
				"2024-10-20 22:00:02 100 239.2 4 [OL]",
				"2024-10-20 22:00:04 100 239.2 4 [OL]",
			},
			want: 0,
		},
		{
			name: "fast poweroff - restores well under the 90-line/3min threshold",
			lines: []string{
				"2024-10-20 23:45:55 100 239.2 4 [OL]",
				"2024-10-20 23:45:57 100 239.5 6 [OB]",
				"2024-10-20 23:45:59 100 239.5 6 [OB]",
				"2024-10-20 23:46:01 100 239.5 6 [OB]",
				"2024-10-20 23:46:17 100 239.2 4 [OL]",
			},
			want: 3,
		},
		{
			name:  "sustained poweroff - reaches the 90-line/3min threshold",
			lines: onBatteryLines(90),
			want:  90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longestOnBatteryStreak(tt.lines); got != tt.want {
				t.Errorf("longestOnBatteryStreak() = %d, want %d", got, tt.want)
			}
		})
	}
}

// longestOnBatteryStreak returns the longest run of consecutive [OB] lines,
// the same signal AlertFastPowerOff uses (90 lines, at upslog's 2s interval,
// is the ~3 minute sustained-outage threshold).
func longestOnBatteryStreak(lines []string) int {
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
