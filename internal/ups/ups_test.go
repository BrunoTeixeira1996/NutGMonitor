package ups

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

func TestMain(m *testing.M) {
	if err := logger.Setup("logs"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

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

// readLogFile reads a whole fixture log (unlike getLogLines, which only
// looks at the last minute relative to the current time - fixtures have
// fixed historical timestamps, so we just read every line here).
func readLogFile(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	var lines []string
	for _, l := range strings.Split(string(data), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// longestOnBatteryStreak returns the longest run of consecutive [OB] lines
// in a log, which is the same signal AlertFastPowerOff uses (90 lines, at
// upslog's 2s interval, is the ~3 minute sustained-outage threshold).
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

func TestLog_Normal(t *testing.T) {
	lines := readLogFile(t, filepath.Join("testdata", "log_normal.txt"))

	if streak := longestOnBatteryStreak(lines); streak != 0 {
		t.Errorf("expected no [OB] lines in a normal log, got a streak of %d", streak)
	}
}

func TestLog_FastPoweroff(t *testing.T) {
	lines := readLogFile(t, filepath.Join("testdata", "log_fast_poweroff.txt"))

	streak := longestOnBatteryStreak(lines)
	if streak == 0 {
		t.Fatal("expected some [OB] lines in a fast-poweroff log")
	}
	if streak >= 90 {
		t.Errorf("expected a fast poweroff (restored under the 90-line/3-minute threshold), got a streak of %d", streak)
	}
}

func TestLog_Poweroff(t *testing.T) {
	lines := readLogFile(t, filepath.Join("testdata", "log_poweroff.txt"))

	if streak := longestOnBatteryStreak(lines); streak < 90 {
		t.Errorf("expected a sustained poweroff (at or over the 90-line/3-minute threshold), got a streak of %d", streak)
	}
}
