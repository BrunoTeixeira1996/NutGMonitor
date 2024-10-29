package targets

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

// Proxmox adheres to the standard Linux shutdown mechanism. It utilizes a script located in /root called off (which runs the command halt -f -f -p). This script terminates all running processes and powers down the machine. The shutdown process is initiated via SSH using ssh -i <key> <target>. Additionally, the target system's authorized_keys file contains the following restriction: no-pty,no-X11-forwarding,command='sudo /root/off' ssh-rsa .... This setup ensures that the SSH connection can only execute the off script and nothing else.

func ShutdownProxmox(sshkey, ip string) error {
	// Set a timeout for the SSH command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", "-i", sshkey, "root@"+ip)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Log.Printf("[shutdown proxmox info] timeout while trying to shutdown proxmox: %s - proxmox should be down by now\n", err)
			return nil
		}
		return fmt.Errorf("[shutdown proxmox error] while powering off proxmox: %s (%s)\n", string(output), err)
	}
	return nil
}
