Following the initial setup of the UPS, our objective is to continuously monitor its performance and automatically shut down connected devices when needed.

**NutGMonitor** is running on a Rasperry Pi

Blog post [here](https://brunoteixeira1996.github.io/posts/2024-10-29-automating-my-ups/)

# Setup

Utilizing Network UPS Tools ([NUT](https://networkupstools.org/)) within a [Docker container](https://github.com/instantlinux/docker-tools/blob/main/images/nut-upsd/README.md) to monitor UPS statistics.

Employing Prometheus Alert Manager to receive real-time updates on UPS status from NUT.

Implementing NutGmonitor in a Docker container to manage UPS actions during power loss.


# Action

When the UPS stops receiving power, it triggers a warning to the Prometheus Alert Manager after 2 minutes. The Alert Manager then forwards this information to NutGMonitor.

If the power loss continues for approximately 5 minutes, NutGMonitor initiates the shutdown of devices connected to the UPS.

## Power off mechanism

The power-off mechanism is tailored to the target system. For instance, if we're working with a Linux machine, we create an `off` bash script containing `halt -f -f -p`. Next, we generate an SSH key for connecting from the NutGMonitor instance to the target. In the target's `authorized_keys`, we restrict the key's access with `no-pty,no-X11-forwarding,command="sudo /root/off"` to ensure it can only execute the off script.

NutGMonitor is also compatible with [Gokrazy](https://gokrazy.org/). Since Gokrazy includes only a limited set of Unix utilities, we can initiate a shutdown by sending a POST request to a specific target.

Once all targets are powered down, we then shut down the Raspberry Pi running NutGMonitor, disconnecting it from the UPS.

All relevant information is forwarded to a Telegram bot and sent via email for notifications.

## Fast Power Loss

Occasionally, a quick power loss followed by a restoration can occur before the 2-minute threshold is met, preventing the Alert Manager from receiving any updates. To address this, I monitor these power off/on events using [upslog](https://networkupstools.org/docs/man/upslog.html) within the NUT Docker container, which I've integrated into the modified `entrypoint.sh` script by adding the following:

```bash
/usr/bin/upslog -u $USER -s ups -i 2 -l /var/log/upslog.txt -f "%TIME @Y-@m-@d @H:@M:@S% %VAR battery.charge% %VAR input.voltage% %VAR ups.load% [%VAR ups.status%]"
```
