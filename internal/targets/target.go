package targets

import (
	"fmt"
	"os"
	"os/exec"
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
			IP:           "192.168.30.12",
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
			IP:           "192.168.30.200",
			SSHKey:       currentPath + "/id_ed25519_nas1",
			ShutdownFunc: ShutdownNas,
		},
	}
}

func (t *Target) ValidateSSHKeys() error {
	if t.SSHKey != "" {
		_, err := os.Stat(t.SSHKey)
		if os.IsNotExist(err) {
			return fmt.Errorf("[target error] %s ssh key does not exist in path: %s\n", t.SSHKey, err)
		}
	}
	return nil
}

func ShutdownTargets(targets []Target) {
	logger.Log.Println("[targets info] preparing to shutdown the following targets:", targets)
	for _, t := range targets {
		if t.Name == "pinute" {
			// pinute is the last target to get shutdown
			continue
		}

		logger.Log.Printf("[targets info] powering off %s ...\n", t.Name)
		if err := t.ShutdownFunc(t.SSHKey, t.IP); err != nil {
			logger.Log.Printf("[targets error] could not shutdown %s: %s\n", t.Name, err)
		} else {
			logger.Log.Printf("[targets info] target %s was shut down\n", t.Name)
		}
	}
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
			if t.Name == "nas1" {
				// ignore nas1 because this will be off at the moment of this
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
