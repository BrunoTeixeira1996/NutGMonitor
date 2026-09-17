package targets

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

type Target struct {
	Name         string
	IP           string
	SSHKey       string
	ShutdownFunc func(sshkey, ip string) error
}

func InitTargets() []Target {
	currentPath, err := os.Getwd()
	if err != nil {
		logger.Log.Printf("[target] could not get current path: %s\n", err)
		return []Target{}
	}
	return []Target{
		{
			Name:         "gokrazy",
			IP:           "192.168.30.13:1080",
			SSHKey:       "",
			ShutdownFunc: ShutdownGokrazy,
		},
		{
			Name:         "proxmox",
			IP:           "192.168.30.3",
			SSHKey:       currentPath + "/id_rsa_proxmox",
			ShutdownFunc: ShutdownProxmox,
		},
		{
			Name:         "nas1",
			IP:           "192.168.30.15",
			SSHKey:       currentPath + "/id_ed25519_nas1",
			ShutdownFunc: ShutdownNas,
		},
		{
			Name:         "pinute",
			IP:           "192.168.30.14",
			SSHKey:       currentPath + "/id_ed25519_pinute",
			ShutdownFunc: ShutdownPinute,
		},
	}
}

func (t *Target) ValidateSSHKeys() error {
	if t.SSHKey != "" {
		if _, err := os.Stat(t.SSHKey); err != nil {
			return fmt.Errorf("[target error] %s ssh key is not accessible: %s\n", t.SSHKey, err)
		}
	}
	return nil
}

// ShutdownTargets powers off every target except "pinute" (which is always
// shut down last, separately, once everything else is confirmed down).
// Targets are shut down concurrently so that one slow/unresponsive target
// (e.g. gokrazy's HTTP call, which budgets up to 4 minutes) doesn't delay the
// shutdown attempt for every other target queued behind it.
func ShutdownTargets(targets []Target) {
	logger.Log.Println("[targets info] preparing to shutdown the following targets:", targets)

	var wg sync.WaitGroup
	for _, t := range targets {
		if t.Name == "pinute" {
			// pinute is the last target to get shutdown
			continue
		}

		wg.Add(1)
		go func(t Target) {
			defer wg.Done()

			logger.Log.Printf("[targets info] powering off %s ...\n", t.Name)
			if err := t.ShutdownFunc(t.SSHKey, t.IP); err != nil {
				// Many shutdown commands (halt, shutdown -P) terminate the
				// SSH session mid-flight, so a non-nil error here is common
				// even when the target actually powered off successfully.
				// CheckTargetsStatus() is the authoritative signal for that.
				logger.Log.Printf("[targets info] shutdown command for %s returned an error (often expected if the command closes the connection): %s\n", t.Name, err)
			} else {
				logger.Log.Printf("[targets info] target %s was shut down\n", t.Name)
			}
		}(t)
	}
	wg.Wait()
}

func isTargetAlive(ip, name string) bool {
	cmd := exec.Command("ping", "-c", "1", "-W", "2", ip)
	err := cmd.Run()

	if err == nil {
		logger.Log.Printf("[target info] %s is still alive ...\n", name)
		return true
	}

	logger.Log.Printf("[target info] %s is not alive anymore: %s\n", name, err)
	return false
}

// CheckTargetsStatus checks the status of the given targets and returns a list of down targets.
func CheckTargetsStatus(targets []Target) []string {
	checkInterval := time.Second * 10 // How often to check the targets
	startTime := time.Now()
	downTargets := []string{}

	logger.Log.Println("[targets info] starting to check targets status (it will check during 4 minutes) ...")

	for {
		// Check if we've reached the timeout
		if time.Since(startTime) > 4*time.Minute {
			logger.Log.Println("[targets info] timeout reached, stopping checks ...")
			return downTargets // Return down targets, even if empty
		}

		allDown := true // Assume all are down initially

		for _, t := range targets {
			if t.Name == "nas1" || t.Name == "pinute" {
				// ignore nas1 (nas1 will be off at the moment of this) and pinute
				continue
			}

			isAlive := isTargetAlive(t.IP, t.Name) // Check if the target is alive
			if isAlive {
				allDown = false // At least one target is alive
			} else {
				downTargets = append(downTargets, t.Name)
			}
		}

		// Check if all targets are down after finishing the for loop
		if allDown {
			logger.Log.Println("[targets info] all targets are down, stopping checks.")
			return downTargets
		}

		// Wait for the next check
		time.Sleep(checkInterval)
	}
}
