package ups

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/forward"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/targets"
)

type AlertLog struct {
	History []string
}

// Add entry to history slice
func (a *AlertLog) setHistory(logLine string) {
	a.History = append(a.History, logLine)
}

// Removes duplicate lines in the history slice
func (a *AlertLog) cleanHistory() {
	seen := make(map[string]struct{})
	uniqueLines := []string{}

	for _, line := range a.History {
		if _, exists := seen[line]; !exists {
			seen[line] = struct{}{}                 // Mark this line as seen
			uniqueLines = append(uniqueLines, line) // Add to unique lines
		}
	}

	a.History = uniqueLines
}

func (a *AlertLog) displayHistory() {
	logger.Log.Println("[ups info] display history")
	logger.Log.Println("========================================")
	for _, entry := range a.History {
		logger.Log.Println(entry)
	}
	logger.Log.Println("========================================")
}

func (a *AlertLog) getDateFromHistory(pos int) string {
	// inputs 2024-10-20 23:46:03 100 239.2 6 [OB]
	// returns 2024-10-20 23:46:03
	s := strings.Split(a.History[pos], " ")
	return fmt.Sprintf("%s %s", s[0], s[1])
}

func isUPSOnBattery(logLine string) bool {
	return strings.Split(logLine, " ")[5] == "[OB]"
}

// Read log lines and grab the previous minute compared with the current
// time
func getLogLines(logFile string) ([]string, error) {
	var logLines []string

	now := time.Now()
	previousMinute := now.Add(-1 * time.Minute).Format("2006-01-02 15:04")

	file, err := os.Open(logFile)
	if err != nil {
		return []string{}, fmt.Errorf("[alert fast power off error] could not open file %s: %s\n", logFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue // Skip malformed lines
		}

		logTimeStr := parts[0] + " " + parts[1]
		logTime, _ := time.Parse("2006-01-02 15:04:05", logTimeStr)

		if logTime.Format("2006-01-02 15:04") == previousMinute {
			logLines = append(logLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{}, fmt.Errorf("[alert fast power off error] reading file %s: %s\n", logFile, err)
	}

	return logLines, nil

}

// Validates if UPS changed the state from OL to OB by reading the log file
// from upslog output and alerts what is going on
func AlertFastPowerOff(logFile string, nas1Target targets.Target) {
	isAlerting := false
	var alertLog AlertLog

	logger.Log.Println("[ups info] monitoring fast power off ...")

	for {
		logLines, err := getLogLines(logFile)
		if err != nil {
			logger.Log.Fatal(err)
		}

		isCurrentlyOnBattery := false
		for _, l := range logLines {
			if isUPSOnBattery(l) {
				isCurrentlyOnBattery = true
				alertLog.setHistory(l)

				if !isAlerting {
					isAlerting = true
					r := fmt.Sprintf("UPS is now on battery mode - Started at %v\n", alertLog.getDateFromHistory(0))
					logger.Log.Printf("[ups info] %s\n", r)
					forward.ForwardMessage("FAST POWEROFF", r, nil)

				}
			}
		}

		if isCurrentlyOnBattery {
			// 3 min * 60 = 180 secs
			// 180 / 2 (upslog sends log every 2 seconds) = 90
			// meaning the ups is down for at least 3 minutes
			if len(alertLog.History) >= 90 {
				logger.Log.Println("[ups info] alert: nutgmonitor will probably start to turn off devices because 5 minutes passed...")
				alertLog.cleanHistory()
				alertLog.displayHistory()
				break
			}
		} else {
			if isAlerting {
				isAlerting = false
				r := fmt.Sprintf("UPS is no longer on battery mode - Ended at %v\n", alertLog.getDateFromHistory(len(alertLog.History)-1))

				logger.Log.Printf("[ups info] %s\n", r)
				forward.ForwardMessage("FAST POWEROFF", r, nil)

				// since the power cameback we need to turn off nas because WoL will turn on
				// nas everytime the power comesback
				if err := nas1Target.ShutdownFunc(nas1Target.SSHKey, nas1Target.IP); err != nil {
					logger.Log.Printf("[ups error] could not shutdown nas1: %s\n", err)
				} else {
					logger.Log.Printf("[targets info] target nas1 was shut down\n")
				}

				alertLog.cleanHistory()
				alertLog.displayHistory()
				// Reset alert log
				alertLog = AlertLog{}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}
