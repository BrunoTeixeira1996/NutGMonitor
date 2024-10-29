package targets

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

func ShutdownNas(sshkey, ip string) error {
	// Set a timeout for the SSH command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds timeout
	defer cancel()

	// i use sudo shutdown -P 0 here because i dont have any restriction on .ssh/authorized_keys
	// inside the nas1 meaning that I need to be explicit on the command to execute
	// but I could use the same approach as proxmox if I created a off file inside nas
	// but ill leave it like this
	cmd := exec.CommandContext(ctx, "ssh", "-i", sshkey, "brun0@"+ip, "sudo", "shutdown", "-P", "0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Log.Printf("[shutdown nas info] timeout while trying to shutdown nas: %s - nas should be down by now\n", err)
			return nil
		}
		return fmt.Errorf("[shutdown nas error] while powering off nas: %s (%s)\n", string(output), err)
	}
	return nil
}
