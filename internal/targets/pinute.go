package targets

import (
	"context"
	"os/exec"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

func ShutdownPinute(sshkey, ip string) error {
	// Set a timeout for the SSH command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", "-i", sshkey, "brun0@"+ip)

	_, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log.Printf("[shutdown pinute error]: %s\n", err)
	}

	return nil
}
